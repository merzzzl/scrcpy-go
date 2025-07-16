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
	client, err := scrcpy.New("127.0.0.1:5555")
	if err != nil {
		log.Printf("connect: %v", err)

		return
	}

	dec, err := NewDecoder()
	if err != nil {
		log.Printf("decoder: %v", err)

		return
	}

	client.SetVideoHandler(func(frame []byte) {
		_ = dec.Decode(frame)
	})

	ctx, cancel := context.WithCancel(context.Background())
	eg, egctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		err := dec.Looper(egctx)
		if err != nil {
			return fmt.Errorf("decoder looper: %w", err)
		}

		return nil
	})

	eg.Go(func() error {
		err := client.Run(egctx)
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
