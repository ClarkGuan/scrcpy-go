package scrcpy

import "github.com/veandco/go-sdl2/sdl"

func loadImage(renderer *sdl.Renderer, path string) (*sdl.Texture, error) {
	surface, err := sdl.LoadBMP(path)
	if err != nil {
		return nil, err
	}
	defer surface.Free()
	return renderer.CreateTextureFromSurface(surface)
}
