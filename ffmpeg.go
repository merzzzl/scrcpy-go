package scrcpy

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
)

type FFmpeg struct {
	stdin  io.WriteCloser
	stdout io.ReadCloser
	cmd    *exec.Cmd
}

func NewDecoder(ctx context.Context) (*FFmpeg, error) {
	cmd := exec.CommandContext(ctx,
		"ffmpeg",
		"-loglevel", "quiet",
		"-f", "h264",
		"-i", "pipe:0",
		"-pix_fmt", "bgr24",
		"-f", "rawvideo",
		"pipe:1",
	)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("open stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("open stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start ffmpeg: %w", err)
	}

	return &FFmpeg{
		stdin:  stdin,
		stdout: stdout,
		cmd:    cmd,
	}, nil
}

func (f *FFmpeg) VideoHandler(r io.Reader) error {
	_, err := io.Copy(f.stdin, r)
	_ = f.stdin.Close()

	if errors.Is(err, io.EOF) {
		return nil
	}

	return fmt.Errorf("copy buff: %w", err)
}

func (f *FFmpeg) Read(p []byte) (n int, err error) {
	return f.stdout.Read(p)
}

func (f *FFmpeg) Close() error {
	var errs []error

	if f.stdin != nil {
		if err := f.stdin.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if f.stdout != nil {
		if err := f.stdout.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if f.cmd != nil && f.cmd.Process != nil {
		if err := f.cmd.Process.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
			errs = append(errs, err)
		}

		_ = f.cmd.Wait()
	}

	if len(errs) == 0 {
		return nil
	}

	return fmt.Errorf("ffmpeg: %w", errors.Join(errs...))
}
