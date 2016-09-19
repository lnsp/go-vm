package vm

import (
	"time"

	termbox "github.com/nsf/termbox-go"
)

type Interrupt struct {
	Type uint16
	Code uint16
}

type Display interface {
	Draw(width, height int, data []byte)
	Init() error
	Close()
}

type TermboxDisplay struct{}

func (TermboxDisplay) Init() error {
	err := termbox.Init()
	if err != nil {
		return err
	}
	go func() {
		for {
			termbox.PollEvent()
		}
	}()
	return nil
}

func (TermboxDisplay) Close() {
	termbox.Close()
}

type DemoDisplay struct {
	TermboxDisplay
}

func (DemoDisplay) Draw(width, height int, data []byte) {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			termbox.SetCell(x, y, ' ', termbox.ColorDefault, termbox.Attribute((x%8)+1))
		}
	}
	termbox.Flush()
}

type TextDisplay struct {
	TermboxDisplay
}

func (TextDisplay) Draw(width, height int, data []byte) {
	var addr, value uint16
	var char rune

	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			addr = uint16(y*width+x) * WORD_SIZE
			value = ByteOrder.Uint16(data[addr : addr+WORD_SIZE])
			if value == 0 {
				char = ' '
			} else {
				char = rune(value)
			}
			termbox.SetCell(x, y, char, termbox.ColorDefault, termbox.ColorDefault)
		}
	}

	termbox.Flush()
	time.Sleep(50 * time.Millisecond)
}
