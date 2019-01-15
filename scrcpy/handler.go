package scrcpy

import "github.com/veandco/go-sdl2/sdl"

type controlHandler struct {
	controller Controller
	set        mouseEventSet
}

func newControlHandler(controller Controller) *controlHandler {
	ch := controlHandler{controller: controller}
	controller.Register(&ch)
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

	sme := singleMouseEvent{}
	sme.action = AMOTION_EVENT_ACTION_MOVE
	sme.id = 0
	sme.point.x = uint16(event.X)
	sme.point.y = uint16(event.Y)
	return true, ch.controller.PushEvent(&sme)
}

func (ch *controlHandler) handleMouseButtonDown(event *sdl.MouseButtonEvent) (bool, error) {
	if sdl.GetRelativeMouseMode() {

	} else {
		sme := singleMouseEvent{}
		sme.action = AMOTION_EVENT_ACTION_DOWN
		sme.id = 0
		sme.point.x = uint16(event.X)
		sme.point.y = uint16(event.Y)
		return true, ch.controller.PushEvent(&sme)
	}
	return false, nil
}

func (ch *controlHandler) handleMouseButtonUp(event *sdl.MouseButtonEvent) (bool, error) {
	if sdl.GetRelativeMouseMode() {

	} else {
		sme := singleMouseEvent{}
		sme.action = AMOTION_EVENT_ACTION_UP
		sme.id = 0
		sme.point.x = uint16(event.X)
		sme.point.y = uint16(event.Y)
		return true, ch.controller.PushEvent(&sme)
	}
	return false, nil
}

func (ch *controlHandler) handleKeyDown(event *sdl.KeyboardEvent) (bool, error) {
	return false, nil
}

func (ch *controlHandler) handleKeyUp(event *sdl.KeyboardEvent) (bool, error) {
	return false, nil
}
