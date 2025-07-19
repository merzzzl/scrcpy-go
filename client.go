package scrcpy

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"golang.org/x/sync/errgroup"
)

type VideoHandler func(io.Reader) error
type ControlHandler func(context.Context, ControlMessage) error

type ControlMessage struct {
	Type    DeviceMessageType
	Payload []byte
}

type Handshake struct {
	DeviceName string
	CodecID    uint32
	Width      uint32
	Height     uint32
}

type Client struct {
	handshake      Handshake
	videoConn      net.Conn
	controlConn    net.Conn
	videoHandler   VideoHandler
	controlHandler ControlHandler
}

func Dial(ctx context.Context, addr string) (*Client, error) {
	dial := (&net.Dialer{Timeout: 5 * time.Second}).DialContext

	vConn, err := dial(ctx, "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("video dial: %w", err)
	}

	cConn, err := dial(ctx, "tcp", addr)
	if err != nil {
		_ = vConn.Close()

		return nil, fmt.Errorf("control dial: %w", err)
	}

	if tcp, ok := cConn.(*net.TCPConn); ok {
		_ = tcp.SetNoDelay(true)
	}

	if err := readExactly(vConn, dummyLen, nil); err != nil {
		_ = vConn.Close()

		return nil, fmt.Errorf("handshake dummy: %w", err)
	}

	nameRaw := make([]byte, deviceNameLen)

	if err := readExactly(vConn, deviceNameLen, nameRaw); err != nil {
		_ = vConn.Close()

		return nil, fmt.Errorf("handshake name: %w", err)
	}

	meta := make([]byte, videoHeaderLen)

	if err := readExactly(vConn, videoHeaderLen, meta); err != nil {
		_ = vConn.Close()

		return nil, fmt.Errorf("handshake meta: %w", err)
	}

	hs := Handshake{
		DeviceName: string(nameRaw),
		CodecID:    binary.BigEndian.Uint32(meta[:4]),
		Width:      binary.BigEndian.Uint32(meta[4:8]),
		Height:     binary.BigEndian.Uint32(meta[8:12]),
	}

	return &Client{
		videoConn:   vConn,
		controlConn: cConn,
		handshake:   hs,
	}, nil
}

func (c *Client) SetVideoHandler(h VideoHandler) { c.videoHandler = h }

func (c *Client) SetControlHandler(h ControlHandler) { c.controlHandler = h }

func (c *Client) GetHandshake() Handshake { return c.handshake }

func (c *Client) Serve(ctx context.Context) error {
	eg, gctx := errgroup.WithContext(ctx)
	gctx, cancel := context.WithCancel(gctx)

	eg.Go(func() error {
		defer cancel()

		return c.readVideo(gctx)
	})

	eg.Go(func() error {
		defer cancel()

		return c.readControl(gctx)
	})

	return eg.Wait()
}

func (c *Client) readVideo(ctx context.Context) error {
	hdr := make([]byte, 12)
	pr, pw := io.Pipe()
	eg, gctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		if c.videoHandler == nil {
			return nil
		}

		if err := c.videoHandler(pr); err != nil {
			return fmt.Errorf("video handler: %w", err)
		}

		return nil
	})

	eg.Go(func() error {
		for gctx.Err() == nil {
			if err := readWithDeadline(c.videoConn, hdr); err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}

				return fmt.Errorf("read video header: %w", err)
			}

			size := binary.BigEndian.Uint32(hdr[8:12])
			if size == 0 {
				continue
			}

			if _, err := io.CopyN(pw, c.videoConn, int64(size)); err != nil {
				return fmt.Errorf("read video: %w", err)
			}
		}

		return nil
	})

	eg.Go(func() error {
		<-gctx.Done()

		return errors.Join(pw.Close(), pr.Close())
	})

	return eg.Wait()
}

func (c *Client) readControl(ctx context.Context) error {
	for ctx.Err() == nil {
		header := make([]byte, 1)

		if err := readWithDeadline(c.controlConn, header); err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}

			return fmt.Errorf("read control header: %w", err)
		}

		buf := make([]byte, 4096)

		n, err := c.controlConn.Read(buf)
		if err != nil && !errors.Is(err, io.EOF) {
			return fmt.Errorf("read control: %w", err)
		}

		if c.controlHandler != nil {
			if err := c.controlHandler(ctx, ControlMessage{
				Type:    DeviceMessageType(header[0]),
				Payload: append([]byte(nil), buf[:n]...),
			}); err != nil {
				return fmt.Errorf("control handler: %w", err)
			}
		}
	}

	return nil
}

func readWithDeadline(conn net.Conn, b []byte) error {
	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	return readExactly(conn, len(b), b)
}

func readExactly(r io.Reader, n int, buf []byte) error {
	if buf == nil {
		buf = make([]byte, n)
	}

	_, err := io.ReadFull(r, buf)

	return err
}
