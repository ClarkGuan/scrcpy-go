package scrcpy

import (
	"github.com/veandco/go-sdl2/sdl"
)

type SdlEventHandler interface {
	HandleSdlEvent(event sdl.Event) (bool, error)
}

type SdlEventLooper interface {
	Loop() error
	Register(h SdlEventHandler)
	Remove(h SdlEventHandler)
}

type defaultLooper struct {
	handles []SdlEventHandler
}

func NewSdlEventLooper() SdlEventLooper {
	return &defaultLooper{}
}

func (dl *defaultLooper) Register(h SdlEventHandler) {
	dl.handles = append(dl.handles, h)
}

func (dl *defaultLooper) Remove(h SdlEventHandler) {
	for i := range dl.handles {
		if dl.handles[i] == h {
			dl.handles = append(dl.handles[:i], dl.handles[i+1:]...)
			return
		}
	}
}

func (dl *defaultLooper) Loop() error {
	var ev sdl.Event
	for ev = sdl.WaitEvent(); ev != nil; ev = sdl.WaitEvent() {
		if ev.GetType() == sdl.QUIT {
			break
		}

		var b bool
		var err error
		for _, h := range dl.handles {
			if b, err = h.HandleSdlEvent(ev); err != nil {
				return err
			}
			if b {
				break
			}
		}
	}

	return nil
}
