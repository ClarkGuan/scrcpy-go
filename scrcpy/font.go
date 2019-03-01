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

type TextTexture struct {
	text    string
	texture sdl.Texture
}

func (tt *TextTexture) Update(renderer sdl.Renderer, f *Font, text string, color sdl.Color, src *sdl.Rect) error {
	if tt.text == text {
		tt.getTextureSize(src)
		return nil
	}

	tt.text = text
	surface, err := f.GetTextSurface(text, color)
	if err != nil {
		return err
	}
	tt.texture, err = renderer.CreateTextureFromSurface(surface)
	if err == nil {
		tt.getTextureSize(src)
	}
	return err
}

func (tt *TextTexture) Render(renderer sdl.Renderer, dst *sdl.Rect) error {
	if tt.texture == 0 {
		return nil
	}
	return renderer.Copy(tt.texture, nil, dst)
}

func (tt *TextTexture) getTextureSize(src *sdl.Rect) {
	if src != nil {
		_, _, src.W, src.H, _ = tt.texture.Query()
	}
}
