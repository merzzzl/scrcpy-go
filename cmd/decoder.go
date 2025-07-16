package main

import (
	"bufio"
	"bytes"
	"context"
	"image"
	"image/jpeg"
	"io"
	"os/exec"
)

type Decoder struct {
	cmd     *exec.Cmd
	stdin   io.WriteCloser
	decoder *bufio.Reader
	image   image.Image
}

const imgsep = 0xFF

func NewDecoder() (*Decoder, error) {
	cmd := exec.Command("ffmpeg",
		"-hide_banner", "-loglevel", "error",
		"-f", "h264", "-i", "pipe:0",
		"-vf", "fps=30",
		"-c:v", "mjpeg", "-f", "mjpeg", "pipe:1",
	)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return &Decoder{
		cmd:     cmd,
		stdin:   stdin,
		decoder: bufio.NewReader(stdout),
	}, nil
}

func (d *Decoder) Decode(frame []byte) error {
	_, err := d.stdin.Write(frame)

	return err
}

func (d *Decoder) Looper(ctx context.Context) error {
	for ctx.Err() == nil {
		var buf bytes.Buffer

		for {
			b, err := d.decoder.ReadByte()
			if err != nil {
				return err
			}

			if b == imgsep {
				next, _ := d.decoder.Peek(1)

				if len(next) == 1 && next[0] == 0xD8 {
					_ = buf.WriteByte(b)

					b2, _ := d.decoder.ReadByte()

					_ = buf.WriteByte(b2)

					break
				}
			}
		}

		for {
			b, err := d.decoder.ReadByte()
			if err != nil {
				return err
			}

			_ = buf.WriteByte(b)

			if b == imgsep {
				next, _ := d.decoder.Peek(1)

				if len(next) == 1 && next[0] == 0xD9 {
					b2, _ := d.decoder.ReadByte()

					_ = buf.WriteByte(b2)

					break
				}
			}
		}

		img, err := jpeg.Decode(&buf)
		if err != nil {
			return err
		}

		d.image = img
	}

	return nil
}

func (d *Decoder) Image() image.Image {
	return d.image
}

func (d *Decoder) Close() {
	_ = d.stdin.Close()
}
