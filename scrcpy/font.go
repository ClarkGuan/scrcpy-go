package scrcpy

import (
	"github.com/ClarkGuan/go-sdl2/sdl"
	"github.com/ClarkGuan/go-sdl2/ttf"
)

type Font struct {
	f *ttf.Font
}

func OpenFont(path string, size int) (*Font, error) {
	if !ttf.WasInit() {
		if err := ttf.Init(); err != nil {
			return nil, err
		}
	}
	if f, err := ttf.OpenFont(path, size); err != nil {
		return nil, err
	} else {
		f.SetStyle(ttf.STYLE_BOLD)
		return &Font{f}, nil
	}
}

func (f *Font) GetTextSurface(text string, color sdl.Color) (*sdl.Surface, error) {
	return f.f.RenderUTF8Blended(text, color)
}

func (f *Font) GetTextSize(text string) (int, int, error) {
	return f.f.SizeUTF8(text)
}
