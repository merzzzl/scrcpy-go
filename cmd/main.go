package main

import (
	"context"
	"log"

	scrcpy "github.com/merzzzl/scrcpy-go"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, err := scrcpy.Dial(ctx, "127.0.0.1:10000")
	if err != nil {
		log.Printf("connect: %v", err)

		return
	}

	device := client.GetHandshake()
	log.Printf("Connected to %s (%dx%d, codec=%d)\n", device.DeviceName, device.Width, device.Height, device.CodecID)

	dec, err := scrcpy.NewDecoder(ctx)
	if err != nil {
		log.Printf("decoder: %v", err)

		return
	}

	defer func() {
		if err := dec.Close(); err != nil {
			log.Printf("decoder: %v", err)
		}
	}()

	client.SetVideoHandler(dec.VideoHandler)

	go func() {
		err := client.Serve(ctx)
		if err != nil {
			log.Printf("scrcpy client: %v", err)
		}
	}()

	if err := AppUI(ctx, client, dec); err != nil {
		log.Printf("ui: %v", err)
	}
}
