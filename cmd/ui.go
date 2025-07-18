package main

import (
	"context"
	"fmt"
	"image"
	"sync/atomic"
	"time"

	"github.com/gdamore/tcell/v2"
	scrcpy "github.com/merzzzl/scrcpy-go"
	"github.com/qeesung/image2ascii/convert"
)

type StateUI struct {
	screenX      atomic.Uint32
	screenY      atomic.Uint32
	client       *scrcpy.Client
	screen       tcell.Screen
	endOfScreenY atomic.Uint32

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
	waitCommand := false

	for ctx.Err() == nil {
		ev := s.screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			s.screen.Clear()
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyEscape || ev.Rune() == 'q' {
				return
			}

			if ev.Key() == tcell.KeyETX {
				waitCommand = true

				continue
			}

			if !waitCommand {
				continue
			}

			waitCommand = false

			if ev.Rune() == 'n' {
				_ = s.client.ExpandNotificationPanel()
			}

			if ev.Rune() == 's' {
				_ = s.client.ExpandSettingsPanel()
			}

			if ev.Rune() == 'p' {
				_ = s.client.CollapsePanels()
			}

			if ev.Rune() == 'b' {
				_ = s.client.BackOrScreenOn(scrcpy.ActionDown)
				_ = s.client.BackOrScreenOn(scrcpy.ActionUp)
			}

			if ev.Rune() == 'r' {
				_ = s.client.RotateDevice()
			}

			s.event2tcell(fmt.Sprintf("Key: %v, Rune: %q, Modifiers: %v\n", ev.Key(), ev.Rune(), ev.Modifiers()))

		case *tcell.EventMouse:
			w, h := s.screen.Size()
			x, y := ev.Position()

			rx := uint32(float32(s.screenX.Load()) / float32(w) * float32(x))
			ry := uint32(float32(s.screenY.Load()) / float32(h) * float32(y))

			if ev.Buttons() == tcell.Button1 {
				if !s.primaryKeyPressed {
					s.primaryKeyPressed = true
					_ = s.client.InjectTouch(scrcpy.ActionDown, 1, rx, ry, 65535, scrcpy.ButtonPrimary, scrcpy.ButtonPrimary)
				} else {
					_ = s.client.InjectTouch(scrcpy.ActionMove, 1, rx, ry, 65535, 0, scrcpy.ButtonPrimary)
				}

				continue
			}

			if s.primaryKeyPressed {
				s.primaryKeyPressed = false
				_ = s.client.InjectTouch(scrcpy.ActionUp, 1, rx, ry, 65535, scrcpy.ButtonPrimary, 0)
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

	s.endOfScreenY.Store(uint32(len(charMatrix)))
	s.screenX.Store(uint32(screenBounds.Dx()))
	s.screenY.Store(uint32(screenBounds.Dy()))
}

func (s *StateUI) event2tcell(ev string) {
	for x, char := range ev {
		color := tcell.NewRGBColor(int32(255), int32(0), int32(0))

		s.screen.SetContent(x, int(s.endOfScreenY.Load())+1, char, nil, tcell.StyleDefault.Foreground(color))
	}
}
