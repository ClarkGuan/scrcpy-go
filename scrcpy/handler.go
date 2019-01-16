package scrcpy

import "github.com/veandco/go-sdl2/sdl"

const (
	mainPointerKeyCode = 500 + iota
	FireKeyCode
	//FrontKeyCode
	//BackKeyCode
	//LeftKeyCode
	//RightKeyCode
)

type controlHandler struct {
	controller Controller
	set        mouseEventSet

	keyState map[int]*int
	keyMap   map[int]*Point
}

func newControlHandler(controller Controller, keyMap map[int]*Point) *controlHandler {
	ch := controlHandler{controller: controller}
	controller.Register(&ch)
	ch.keyState = make(map[int]*int)
	ch.keyMap = keyMap
	return &ch
}

func (ch *controlHandler) HandleControlEvent(c Controller, ent interface{}) interface{} {
	if sme, ok := ent.(*singleMouseEvent); ok {
		ch.set.accept(sme)
		return &ch.set
	}
	return ent
}

func (ch *controlHandler) HandleSdlEvent(event sdl.Event) (bool, error) {
	switch event.GetType() {
	case sdl.MOUSEMOTION:
		return ch.handleMouseMotion(event.(*sdl.MouseMotionEvent))

	case sdl.MOUSEBUTTONDOWN:
		return ch.handleMouseButtonDown(event.(*sdl.MouseButtonEvent))

	case sdl.MOUSEBUTTONUP:
		return ch.handleMouseButtonUp(event.(*sdl.MouseButtonEvent))

	case sdl.KEYDOWN:
		return ch.handleKeyDown(event.(*sdl.KeyboardEvent))

	case sdl.KEYUP:
		return ch.handleKeyUp(event.(*sdl.KeyboardEvent))
	}

	return false, nil
}

func (ch *controlHandler) handleMouseMotion(event *sdl.MouseMotionEvent) (bool, error) {
	if !sdl.GetRelativeMouseMode() && event.State == 0 {
		return true, nil
	}

	if ch.keyState[mainPointerKeyCode] != nil {
		return ch.sendMouseEvent(AMOTION_EVENT_ACTION_MOVE, *ch.keyState[mainPointerKeyCode], Point{uint16(event.X), uint16(event.Y)})
	} else {
		panic("main pointer state error")
	}
	return true, nil
}

func (ch *controlHandler) handleMouseButtonDown(event *sdl.MouseButtonEvent) (bool, error) {
	if sdl.GetRelativeMouseMode() {

	} else {
		if ch.keyState[mainPointerKeyCode] == nil {
			ch.keyState[mainPointerKeyCode] = fingers.GetId()
			return ch.sendMouseEvent(AMOTION_EVENT_ACTION_DOWN, *ch.keyState[mainPointerKeyCode], Point{uint16(event.X), uint16(event.Y)})
		} else {
			panic("main pointer state error")
		}
	}
	return false, nil
}

func (ch *controlHandler) handleMouseButtonUp(event *sdl.MouseButtonEvent) (bool, error) {
	if sdl.GetRelativeMouseMode() {

	} else {
		if ch.keyState[mainPointerKeyCode] != nil {
			b, e := ch.sendMouseEvent(AMOTION_EVENT_ACTION_UP, *ch.keyState[mainPointerKeyCode], Point{uint16(event.X), uint16(event.Y)})
			fingers.Recycle(ch.keyState[mainPointerKeyCode])
			ch.keyState[mainPointerKeyCode] = nil
			return b, e
		} else {
			panic("main pointer state error")
		}
	}
	return false, nil
}

func (ch *controlHandler) handleKeyDown(event *sdl.KeyboardEvent) (bool, error) {
	alt := event.Keysym.Mod&(sdl.KMOD_RALT|sdl.KMOD_LALT) != 0
	if alt {
		return true, nil
	}
	ctrl := event.Keysym.Mod&(sdl.KMOD_RCTRL|sdl.KMOD_LCTRL) != 0
	if !ctrl {
		keyCode := int(event.Keysym.Sym)
		if poi := ch.keyMap[keyCode]; poi != nil {
			if ch.keyState[keyCode] == nil {
				ch.keyState[keyCode] = fingers.GetId()
				return ch.sendMouseEvent(AMOTION_EVENT_ACTION_DOWN, *ch.keyState[keyCode], *poi)
			} else {
				return ch.sendMouseEvent(AMOTION_EVENT_ACTION_MOVE, *ch.keyState[keyCode], *poi)
			}
		}
	}
	return true, nil
}

func (ch *controlHandler) handleKeyUp(event *sdl.KeyboardEvent) (bool, error) {
	alt := event.Keysym.Mod&(sdl.KMOD_RALT|sdl.KMOD_LALT) != 0
	if alt {
		return true, nil
	}
	ctrl := event.Keysym.Mod&(sdl.KMOD_RCTRL|sdl.KMOD_LCTRL) != 0
	if ctrl && event.Keysym.Sym == sdl.K_x {
		mode := sdl.GetRelativeMouseMode()
		sdl.SetRelativeMouseMode(!mode)
	}

	if !ctrl {
		keyCode := int(event.Keysym.Sym)
		if poi := ch.keyMap[keyCode]; poi != nil {
			b, e := ch.sendMouseEvent(AMOTION_EVENT_ACTION_UP, *ch.keyState[keyCode], *poi)
			fingers.Recycle(ch.keyState[keyCode])
			ch.keyState[keyCode] = nil
			return b, e
		}
	}
	return true, nil
}

func (ch *controlHandler) sendMouseEvent(action androidMotionEventAction, id int, p Point) (bool, error) {
	sme := singleMouseEvent{action: action}
	sme.id = id
	sme.Point = p
	return true, ch.controller.PushEvent(&sme)
}
