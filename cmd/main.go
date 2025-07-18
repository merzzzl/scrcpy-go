package main

import (
	"context"
	"fmt"
	"log"
	"os"

	scrcpy "github.com/merzzzl/scrcpy-go"
	"golang.org/x/sync/errgroup"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	client, err := scrcpy.Dial(ctx, "127.0.0.1:10000")
	if err != nil {
		log.Printf("connect: %v", err)

		return
	}

	device := client.GetHandshake()
	log.Printf("Connected to %s (%dx%d, codec=%d)\n", device.DeviceName, device.Width, device.Height, device.CodecID)

	dec, err := NewDecoder()
	if err != nil {
		log.Printf("decoder: %v", err)

		return
	}

	client.SetVideoHandler(func(frame []byte) {
		_ = dec.Decode(frame)
	})

	eg, egctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		err := dec.Looper(egctx)
		if err != nil {
			return fmt.Errorf("decoder looper: %w", err)
		}

		return nil
	})

	eg.Go(func() error {
		err := client.Serve(egctx)
		if err != nil {
			return fmt.Errorf("scrcpy client: %w", err)
		}

		return nil
	})

	eg.Go(func() error {
		defer cancel()
		defer dec.Close()

		err := AppUI(egctx, client, dec)
		if err != nil {
			return fmt.Errorf("ui: %w", err)
		}

		return nil
	})

	if err := eg.Wait(); err != nil {
		log.Printf("main: %v", err)
	}

	cancel()
	os.Exit(0)
}
