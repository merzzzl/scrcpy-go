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

type VideoHandler func(frame []byte)
type ControlHandler func(ControlMessage)

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
		DeviceName: string(nameRaw[:trimZero(nameRaw)]),
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
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error { return c.readVideo(ctx) })
	eg.Go(func() error { return c.readControl(ctx) })

	return eg.Wait()
}

func (c *Client) readVideo(ctx context.Context) error {
	for ctx.Err() == nil {
		hdr := make([]byte, frameHeaderLen)
		if err := readWithDeadline(c.videoConn, hdr); err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}

			return err
		}

		size := binary.BigEndian.Uint32(hdr[8:12])
		if size == 0 {
			continue
		}

		payload := make([]byte, size)

		if err := readWithDeadline(c.videoConn, payload); err != nil {
			return err
		}

		if c.videoHandler != nil {
			c.videoHandler(payload)
		}
	}

	return nil
}

func (c *Client) readControl(ctx context.Context) error {
	for ctx.Err() == nil {
		header := make([]byte, 1)
		if err := readWithDeadline(c.controlConn, header); err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}

			return err
		}

		buf := make([]byte, 4096)

		n, err := c.controlConn.Read(buf)
		if err != nil && !errors.Is(err, io.EOF) {
			return err
		}

		if c.controlHandler != nil {
			c.controlHandler(ControlMessage{
				Type:    DeviceMessageType(header[0]),
				Payload: append([]byte(nil), buf[:n]...),
			})
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

func trimZero(b []byte) int {
	for i, v := range b {
		if v == 0 {
			return i
		}
	}

	return len(b)
}
