package scrcpy

import (
	"log"

	"github.com/ClarkGuan/go-sdl2/sdl"
)

const (
	displayMargins = 96
)

func sdlInitAndConfigure() (err error) {
	if !sdl.SetHint(sdl.HINT_NO_SIGNAL_HANDLERS, "1") {
		log.Println("Cannot request to keep default signal handlers")
	}

	if err = sdl.Init(sdl.INIT_VIDEO); err != nil {
		return
	}

	if !sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "2") {
		log.Println("Could not enable bilinear filtering")
	}

	if !sdl.SetHint("SDL_MOUSE_FOCUS_CLICKTHROUGH", "1") {
		log.Println("Could not enable mouse focus clickthrough")
	}

	return
}

func maxUint16(a, b uint16) uint16 {
	if a > b {
		return a
	}

	return b
}

func minUnt32(a, b uint32) uint32 {
	if a < b {
		return a
	}

	return b
}

func getPreferredDisplayBounds() (bounds size, err error) {
	if rect, err := sdl.GetDisplayUsableBounds(0); err != nil {
		return bounds, err
	} else {
		bounds.width = maxUint16(0, uint16(rect.W-displayMargins))
		bounds.height = maxUint16(0, uint16(rect.H-displayMargins))
		return bounds, nil
	}
}

func getOptimalSize(currentSize, frameSize size) size {
	if frameSize.width == 0 || frameSize.height == 0 {
		// avoid division by 0
		return currentSize
	}

	var w, h uint32

	if displaySize, err := getPreferredDisplayBounds(); err == nil {
		w = minUnt32(uint32(currentSize.width), uint32(displaySize.width))
		h = minUnt32(uint32(currentSize.height), uint32(displaySize.height))
	} else {
		w = uint32(currentSize.width)
		h = uint32(currentSize.height)
	}

	if uint32(frameSize.width)*h > uint32(frameSize.height)*w {
		h = uint32(frameSize.height) * w / uint32(frameSize.width)
	} else {
		w = uint32(frameSize.width) * h / uint32(frameSize.height)
	}

	return size{width: uint16(w), height: uint16(h)}
}

func getInitialOptimalSize(frameSize size) size {
	return getOptimalSize(frameSize, frameSize)
}

type screen struct {
	window    sdl.Window
	renderer  sdl.Renderer
	texture   sdl.Texture
	frameSize size
	hasFrame  bool
	Renderers []Renderer
	initFlag  bool
}

func (s *screen) InitRendering(deviceName string, frameSize size) (err error) {
	s.frameSize = frameSize
	windowSize := getInitialOptimalSize(frameSize)
	windowFlags := sdl.WINDOW_HIDDEN // SDL_WINDOW_RESIZABLE
	windowFlags |= sdl.WINDOW_ALLOW_HIGHDPI
	if s.window, err = sdl.CreateWindow(deviceName, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, int32(windowSize.width), int32(windowSize.height), uint32(windowFlags)); err != nil {
		return
	}
	if s.renderer, err = sdl.CreateRenderer(s.window, -1, sdl.RENDERER_ACCELERATED); err != nil {
		s.Close()
		return
	}
	if err = s.renderer.SetLogicalSize(int32(frameSize.width), int32(frameSize.height)); err != nil {
		s.Close()
		return
	}
	if debugOpt.Debug() {
		log.Printf("Initial texture: %d, %d", frameSize.width, frameSize.height)
	}
	if err = s.createTexture(frameSize.width, frameSize.height); err != nil {
		s.Close()
		return
	}
	return
}

func (s *screen) Close() error {
	if s.texture != 0 {
		s.texture.Destroy()
		s.texture = 0
	}
	if s.renderer != 0 {
		s.renderer.Destroy()
		s.renderer = 0
	}
	if s.window != 0 {
		s.window.Destroy()
		s.window = 0
	}
	return nil
}

func (s *screen) showWindow() {
	s.window.Show()
}

func (s *screen) prepareForFrame(newFrameSize size) (err error) {
	if s.frameSize.width != newFrameSize.width || s.frameSize.height != newFrameSize.height {
		if err = s.renderer.SetLogicalSize(int32(newFrameSize.width), int32(newFrameSize.height)); err != nil {
			log.Printf("Could not set renderer logical size: %v\n", err)
			return
		}

		s.texture.Destroy()
		w, h := s.window.GetSize()
		currentSize := size{width: uint16(w), height: uint16(h)}
		targetSize := size{width: uint16(uint32(currentSize.width) * uint32(newFrameSize.width) / uint32(s.frameSize.width)),
			height: uint16(uint32(currentSize.height) * uint32(newFrameSize.height) / uint32(s.frameSize.height))}
		targetSize = getOptimalSize(targetSize, newFrameSize)
		s.window.SetSize(int32(targetSize.width), int32(targetSize.height))
		s.frameSize = newFrameSize
		if debugOpt.Debug() {
			log.Printf("New texture: %d, %d\n", newFrameSize.width, newFrameSize.height)
		}
		if err = s.createTexture(newFrameSize.width, newFrameSize.height); err != nil {
			log.Printf("Could not create texture: %v\n", err)
			return
		}
	}

	return
}

func (s *screen) render() {
	if !s.initFlag {
		s.initFlag = true
		for _, r := range s.Renderers {
			r.Init(s.renderer)
		}
	}
	s.renderer.Clear()
	s.renderer.Copy(s.texture, nil, nil)
	for _, r := range s.Renderers {
		r.Render(s.renderer)
	}
	s.renderer.Present()
}

func (s *screen) createTexture(w, h uint16) (err error) {
	// 在 MacOS 上可以创建 NV12 的 Texture 进行硬件加速
	s.texture, err = s.renderer.CreateTexture(sdl.PIXELFORMAT_YV12, sdl.TEXTUREACCESS_STREAMING, int32(w), int32(h))
	return
}

func (s *screen) addRendererFunc(r Renderer) {
	s.Renderers = append(s.Renderers, r)
}

type Renderer interface {
	Init(r sdl.Renderer)
	Render(r sdl.Renderer)
}
