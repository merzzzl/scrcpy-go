package scrcpy

import (
	"context"
	"errors"
	"io"
	"log"
	"net"
	"time"

	"golang.org/x/sync/errgroup"
)

type VideoHandler func([]byte)

type ControlMessage struct {
	Type    ControlMessageType
	Payload []byte
}

type ControlHandler func(ControlMessage)

type Client struct {
	videoConn      net.Conn
	controlConn    net.Conn
	videoHandler   VideoHandler
	controlHandler ControlHandler
}

func New(addr string) (*Client, error) {
	vConn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	cConn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	handshake := make([]byte, lenHandshake)

	_, err = io.ReadFull(vConn, handshake)
	if err != nil {
		return nil, err
	}

	return &Client{
		videoConn:   vConn,
		controlConn: cConn,
	}, nil
}

func (c *Client) SetVideoHandler(h VideoHandler) {
	c.videoHandler = h
}

func (c *Client) SetControlHandler(h ControlHandler) {
	c.controlHandler = h
}

func (c *Client) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		return c.readVideo(ctx)
	})

	eg.Go(func() error {
		return c.readControl(ctx)
	})

	return eg.Wait()
}

func (c *Client) readVideo(ctx context.Context) error {
	buf := make([]byte, videoFrame)

	for ctx.Err() == nil {
		if err := c.videoConn.SetReadDeadline(time.Now().Add(time.Second * 5)); err != nil {
			log.Printf("readVideo error: %v", err)

			return err
		}

		n, err := c.videoConn.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Printf("readVideo error: %v", err)
			}

			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}

			return err
		}

		if n == 0 {
			continue
		}

		if c.videoHandler != nil {
			frame := make([]byte, n)

			copy(frame, buf[:n])
			c.videoHandler(frame)
		}
	}

	return nil
}

func (c *Client) readControl(ctx context.Context) error {
	buf := make([]byte, controlFrame)

	for ctx.Err() == nil {
		if err := c.controlConn.SetReadDeadline(time.Now().Add(time.Second * 5)); err != nil {
			log.Printf("readControl error: %v", err)

			return err
		}

		n, err := c.controlConn.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Printf("readControl error: %v", err)
			}

			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}

			return err
		}

		if n == 0 {
			continue
		}

		if c.controlHandler != nil {
			msg := ControlMessage{
				Type:    ControlMessageType(buf[0]),
				Payload: append([]byte(nil), buf[1:n]...),
			}

			c.controlHandler(msg)
		}
	}

	return nil
}
