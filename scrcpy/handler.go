package scrcpy

import (
	"log"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

const (
	mainPointerKeyCode = 500 + iota
	FireKeyCode
	VisionKeyCode
	FrontKeyCode
	BackKeyCode
)

const mouseAccuracy = .25
const eventVisionEventUp = sdl.USEREVENT + 3
const mouseVisionDelay = time.Millisecond * 500

type controlHandler struct {
	controller Controller
	screen     *screen
	set        mouseEventSet

	keyState map[int]*int
	keyMap   map[int]*Point

	cachePointer Point

	directionController directionController
	timer               map[uint32]*time.Timer
	doubleHit           bool
	*continuousFire
}

func newControlHandler(controller Controller, screen *screen, keyMap map[int]*Point) *controlHandler {
	ch := controlHandler{controller: controller}
	controller.Register(&ch)
	ch.keyState = make(map[int]*int)
	ch.keyMap = keyMap
	ch.screen = screen
	ch.directionController.keyMap = keyMap
	// 默认开启连发模式
	ch.doubleHit = true
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
	case eventVisionEventUp:
		var b bool
		var e error
		if ch.keyState[VisionKeyCode] != nil {
			b, e = ch.sendMouseEvent(AMOTION_EVENT_ACTION_UP, *ch.keyState[VisionKeyCode], ch.cachePointer)
			fingers.Recycle(ch.keyState[VisionKeyCode])
			ch.keyState[VisionKeyCode] = nil
		}
		return b, e

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

func (ch *controlHandler) outside(p *Point) bool {
	ret := false
	minW := uint16(650)
	maxW := uint16(1200)
	if p.X < minW {
		ret = true
		p.X = minW
	} else if p.X > maxW {
		ret = true
		p.X = maxW
	}

	minH := uint16(100)
	maxH := uint16(850)
	if p.Y < minH {
		ret = true
		p.Y = minH
	} else if p.Y > maxH {
		ret = true
		p.Y = maxH
	}

	return ret
}

func fixMouseBlock(x int32) int32 {
	fx := float64(x)
	ret := int32(fx*mouseAccuracy + .5)
	if ret == 0 && x != 0 {
		if x > 0 {
			ret = 1
		} else {
			ret = -1
		}
	}
	return ret
}

func (ch *controlHandler) visionMoving(event *sdl.MouseMotionEvent, delta int) (bool, error) {
	if ch.keyState[VisionKeyCode] == nil {
		ch.keyState[VisionKeyCode] = fingers.GetId()
		ch.cachePointer = Point{950, 450}
		ch.sendEventDelay(eventVisionEventUp, mouseVisionDelay)
		return ch.sendMouseEvent(AMOTION_EVENT_ACTION_DOWN, *ch.keyState[VisionKeyCode], ch.cachePointer)
	} else {
		ch.cachePointer.X = uint16(int32(ch.cachePointer.X) + fixMouseBlock(event.XRel))
		ch.cachePointer.Y = uint16(int32(ch.cachePointer.Y) + fixMouseBlock(event.YRel) + int32(delta))
		if ch.outside(&ch.cachePointer) {
			b, e := ch.sendMouseEvent(AMOTION_EVENT_ACTION_UP, *ch.keyState[VisionKeyCode], ch.cachePointer)
			fingers.Recycle(ch.keyState[VisionKeyCode])
			ch.keyState[VisionKeyCode] = nil
			return b, e
		} else {
			ch.sendEventDelay(eventVisionEventUp, mouseVisionDelay)
			return ch.sendMouseEvent(AMOTION_EVENT_ACTION_MOVE, *ch.keyState[VisionKeyCode], ch.cachePointer)
		}
	}
}

func (ch *controlHandler) handleMouseMotion(event *sdl.MouseMotionEvent) (bool, error) {
	if sdl.GetRelativeMouseMode() {
		if event.State == 0 {
			return ch.visionMoving(event, 0)
		} else {
			if ch.doubleHit {
				if ch.keyState[mainPointerKeyCode] != nil {
					return ch.sendMouseEvent(AMOTION_EVENT_ACTION_MOVE, *ch.keyState[mainPointerKeyCode], Point{uint16(event.X), uint16(event.Y)})
				} else if ch.keyState[FireKeyCode] != nil {
					ch.visionMoving(event, 0)
					b, e := ch.sendMouseEvent(AMOTION_EVENT_ACTION_UP, *ch.keyState[FireKeyCode], *ch.keyMap[FireKeyCode])
					fingers.Recycle(ch.keyState[FireKeyCode])
					ch.keyState[FireKeyCode] = nil
					return b, e
				} else if ch.continuousFire != nil {
					ch.visionMoving(event, 0)
					return true, nil
				}
			} else {
				if ch.keyState[mainPointerKeyCode] != nil {
					return ch.sendMouseEvent(AMOTION_EVENT_ACTION_MOVE, *ch.keyState[mainPointerKeyCode], Point{uint16(event.X), uint16(event.Y)})
				} else if ch.keyState[FireKeyCode] != nil {
					ch.visionMoving(event, 0)
					return ch.sendMouseEvent(AMOTION_EVENT_ACTION_MOVE, *ch.keyState[FireKeyCode], *ch.keyMap[FireKeyCode])
				} else {
					panic("fire pointer state error")
				}
			}
		}
	} else {
		if ch.keyState[VisionKeyCode] != nil {
			b, e := ch.sendMouseEvent(AMOTION_EVENT_ACTION_UP, *ch.keyState[VisionKeyCode], ch.cachePointer)
			fingers.Recycle(ch.keyState[VisionKeyCode])
			ch.keyState[VisionKeyCode] = nil
			return b, e
		}

		if event.State == 0 {
			return true, nil
		}

		if ch.keyState[mainPointerKeyCode] != nil {
			return ch.sendMouseEvent(AMOTION_EVENT_ACTION_MOVE, *ch.keyState[mainPointerKeyCode], Point{uint16(event.X), uint16(event.Y)})
		} else {
			panic("main pointer state error")
		}
	}

	return true, nil
}

func (ch *controlHandler) handleMouseButtonDown(event *sdl.MouseButtonEvent) (bool, error) {
	if sdl.GetRelativeMouseMode() {
		if ch.doubleHit {
			if ch.continuousFire == nil {
				ch.continuousFire = new(continuousFire)
				ch.continuousFire.Point = *ch.keyMap[FireKeyCode]
				ch.continuousFire.Start(ch.controller)
				return true, nil
			}
		} else {
			if ch.keyState[FireKeyCode] == nil {
				ch.keyState[FireKeyCode] = fingers.GetId()
				if debugOpt {
					log.Println("按下开火键")
				}
				return ch.sendMouseEvent(AMOTION_EVENT_ACTION_DOWN, *ch.keyState[FireKeyCode], *ch.keyMap[FireKeyCode])
			}
		}
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
		if ch.keyState[mainPointerKeyCode] != nil {
			b, e := ch.sendMouseEvent(AMOTION_EVENT_ACTION_UP, *ch.keyState[mainPointerKeyCode], Point{uint16(event.X), uint16(event.Y)})
			fingers.Recycle(ch.keyState[mainPointerKeyCode])
			ch.keyState[mainPointerKeyCode] = nil
			return b, e
		} else if ch.keyState[FireKeyCode] != nil {
			b, e := ch.sendMouseEvent(AMOTION_EVENT_ACTION_UP, *ch.keyState[FireKeyCode], *ch.keyMap[FireKeyCode])
			fingers.Recycle(ch.keyState[FireKeyCode])
			ch.keyState[FireKeyCode] = nil
			if debugOpt {
				log.Println("松开开火键")
			}
			return b, e
		} else if ch.continuousFire != nil {
			ch.continuousFire.Stop()
			ch.continuousFire = nil
			return true, nil
		}
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
		} else {
			switch event.Keysym.Sym {
			case sdl.K_w:
				ch.directionController.frontDown()
				return true, ch.directionController.sendMouseEvent(ch.controller)

			case sdl.K_s:
				ch.directionController.backDown()
				return true, ch.directionController.sendMouseEvent(ch.controller)

			case sdl.K_a:
				ch.directionController.leftDown()
				return true, ch.directionController.sendMouseEvent(ch.controller)

			case sdl.K_d:
				ch.directionController.rightDown()
				return true, ch.directionController.sendMouseEvent(ch.controller)
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

	if ctrl {
		switch event.Keysym.Sym {
		case sdl.K_x:
			sdl.SetRelativeMouseMode(!sdl.GetRelativeMouseMode())
		case sdl.K_z:
			ch.doubleHit = !ch.doubleHit
			log.Printf("连击模式:%t\n", ch.doubleHit)
		}
	} else {
		keyCode := int(event.Keysym.Sym)
		if poi := ch.keyMap[keyCode]; poi != nil {
			switch keyCode {
			case sdl.K_m:
				fallthrough
			case sdl.K_t:
				fallthrough
			case sdl.K_TAB:
				sdl.SetRelativeMouseMode(!sdl.GetRelativeMouseMode())
			}

			b, e := ch.sendMouseEvent(AMOTION_EVENT_ACTION_UP, *ch.keyState[keyCode], *poi)
			fingers.Recycle(ch.keyState[keyCode])
			ch.keyState[keyCode] = nil
			return b, e
		} else {
			switch event.Keysym.Sym {
			case sdl.K_w:
				ch.directionController.frontUp()
				return true, ch.directionController.sendMouseEvent(ch.controller)

			case sdl.K_s:
				ch.directionController.backUp()
				return true, ch.directionController.sendMouseEvent(ch.controller)

			case sdl.K_a:
				ch.directionController.leftUp()
				return true, ch.directionController.sendMouseEvent(ch.controller)

			case sdl.K_d:
				ch.directionController.rightUp()
				return true, ch.directionController.sendMouseEvent(ch.controller)

			case sdl.K_k:
				mm := newMirrorMotion(Point{687, 227}, Point{675, 596})
				mm.Start(ch.controller)
				return true, nil

			case sdl.K_l:
				mm := newMirrorMotion(Point{687, 227}, Point{675, 341})
				mm.Start(ch.controller)
				return true, nil
			}
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

func (ch *controlHandler) sendEventDelay(typ uint32, duration time.Duration) {
	if ch.timer == nil {
		ch.timer = make(map[uint32]*time.Timer)
	}

	if ch.timer[typ] != nil {
		ch.timer[typ].Reset(duration)
	} else {
		ch.timer[typ] = time.AfterFunc(duration, func() {
			sdl.PushEvent(&sdl.UserEvent{Type: typ})
		})
	}
}
