package main

import (
	"context"
	"image"
	"io"
	"sync"

	"github.com/gdamore/tcell/v2"
	scrcpy "github.com/merzzzl/scrcpy-go"
	"github.com/qeesung/image2ascii/convert"
)

type StateUI struct {
	mutex  sync.Mutex
	client *scrcpy.Client
	screen tcell.Screen
	width  int
	height int
}

func AppUI(ctx context.Context, client *scrcpy.Client, decoder *scrcpy.FFmpeg) error {
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

	hs := client.GetHandshake()

	state := StateUI{
		client: client,
		screen: screen,
		width:  int(hs.Width),
		height: int(hs.Height),
	}

	go func() {
		buf := make([]byte, state.width*state.height*3)

		for ctx.Err() == nil {
			if _, err := io.ReadFull(decoder, buf); err != nil {
				return
			}

			go state.img2tcell(buf)
		}
	}()

	state.eventsHandler(ctx)

	return nil
}

func (s *StateUI) eventsHandler(ctx context.Context) {
	var primaryKeyPressed bool

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

			rx := uint32(float32(s.width) / float32(w) * float32(x))
			ry := uint32(float32(s.height) / float32(h) * float32(y))

			if ev.Buttons() == tcell.Button1 {
				if !primaryKeyPressed {
					primaryKeyPressed = true
					_ = s.client.InjectTouch(scrcpy.ActionDown, 1, rx, ry, 65535, scrcpy.ButtonPrimary, scrcpy.ButtonPrimary)
				} else {
					_ = s.client.InjectTouch(scrcpy.ActionMove, 1, rx, ry, 65535, 0, scrcpy.ButtonPrimary)
				}

				continue
			}

			if primaryKeyPressed {
				primaryKeyPressed = false
				_ = s.client.InjectTouch(scrcpy.ActionUp, 1, rx, ry, 65535, scrcpy.ButtonPrimary, 0)
			}
		}
	}
}

func (s *StateUI) img2tcell(buf []byte) {
	if !s.mutex.TryLock() {
		return
	}

	defer s.mutex.Unlock()

	img := image.NewNRGBA(image.Rect(0, 0, s.width, s.height))

	for i := 0; i < s.width*s.height; i++ {
		b := buf[3*i+0]
		g := buf[3*i+1]
		r := buf[3*i+2]

		j := 4 * i
		img.Pix[j+0] = r
		img.Pix[j+1] = g
		img.Pix[j+2] = b
		img.Pix[j+3] = 0xff
	}

	converter := convert.NewImageConverter()
	charMatrix := converter.Image2CharPixelMatrix(img, &convert.DefaultOptions)

	for y, row := range charMatrix {
		for x, char := range row {
			color := tcell.NewRGBColor(int32(char.R), int32(char.G), int32(char.B))

			s.screen.SetContent(x, y, rune(char.Char), nil, tcell.StyleDefault.Foreground(color))
		}
	}

	s.screen.Show()
}
