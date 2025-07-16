package main

import (
	"context"
	"image"
	"sync/atomic"
	"time"

	"github.com/gdamore/tcell/v2"
	scrcpy "github.com/merzzzl/scrcpy-go"
	"github.com/qeesung/image2ascii/convert"
)

type StateUI struct {
	screenX atomic.Uint32
	screenY atomic.Uint32
	client  *scrcpy.Client
	screen  tcell.Screen

	primaryKeyPressed bool
}

func AppUI(ctx context.Context, client *scrcpy.Client, decoder *Decoder) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	screen, err := tcell.NewScreen()
	if err != nil {
		return err
	}

	if err := screen.Init(); err != nil {
		return err
	}

	defer screen.Fini()

	screen.EnableMouse()

	state := StateUI{
		client: client,
		screen: screen,
	}

	go func() {
		for ctx.Err() == nil {
			img := decoder.Image()

			state.img2tcell(img)
			time.Sleep(25 * time.Millisecond)
		}
	}()

	state.eventsHandler(ctx)
	cancel()

	return nil
}

func (s *StateUI) eventsHandler(ctx context.Context) {
	for ctx.Err() == nil {
		ev := s.screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			s.screen.Clear()
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyEscape || ev.Rune() == 'q' {
				return
			}
		case *tcell.EventMouse:
			w, h := s.screen.Size()
			x, y := ev.Position()

			rx := uint32(float32(s.screenX.Load()) / float32(w) * float32(x))
			ry := uint32(float32(s.screenY.Load()) / float32(h) * float32(y))

			if ev.Buttons() == tcell.Button1 {
				if !s.primaryKeyPressed {
					s.primaryKeyPressed = true
					_ = s.client.InjectTouch(scrcpy.ActionDown, 1, rx, ry, uint16(s.screenX.Load()), uint16(s.screenY.Load()), 1, 1)
				} else {
					_ = s.client.InjectTouch(scrcpy.ActionMove, 1, rx, ry, uint16(s.screenX.Load()), uint16(s.screenY.Load()), 1, 1)
				}

				continue
			}

			if s.primaryKeyPressed {
				s.primaryKeyPressed = false
				_ = s.client.InjectTouch(scrcpy.ActionUp, 1, rx, ry, uint16(s.screenX.Load()), uint16(s.screenY.Load()), 1, 1)
			}
		}
	}
}

func (s *StateUI) img2tcell(img image.Image) {
	if img == nil {
		return
	}

	screenBounds := img.Bounds()
	converter := convert.NewImageConverter()
	charMatrix := converter.Image2CharPixelMatrix(img, &convert.DefaultOptions)

	for y, row := range charMatrix {
		for x, char := range row {
			color := tcell.NewRGBColor(int32(char.R), int32(char.G), int32(char.B))

			s.screen.SetContent(x, y, rune(char.Char), nil, tcell.StyleDefault.Foreground(color))
		}
	}

	s.screen.Show()

	s.screenX.Store(uint32(screenBounds.Dx()))
	s.screenY.Store(uint32(screenBounds.Dy()))
}
