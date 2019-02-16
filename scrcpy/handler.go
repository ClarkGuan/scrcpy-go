package scrcpy

import (
	"log"
	"path/filepath"
	"time"

	"github.com/ClarkGuan/go-sdl2/sdl"
)

const (
	mainPointerKeyCode = 500 + iota
	FireKeyCode
	VisionKeyCode
	FrontKeyCode
	BackKeyCode
	WheelKeyCode
)

const mouseAccuracy = .085
const mouseVisionDelay = time.Millisecond * 500
const eventVisionEventUp = sdl.USEREVENT + 3
const eventDirectionEvent = sdl.USEREVENT + 4
const eventWheelEvent = sdl.USEREVENT + 5

type controlHandler struct {
	controller Controller
	set        mouseEventSet

	keyState map[int]*int
	keyMap   map[int]*Point

	ctrlKeyState map[int]*int
	ctrlKeyMap   map[int]*Point

	visionCachePointer Point
	wheelCachePointer  Point

	directionController directionController
	timer               map[uint32]*time.Timer
	doubleHit           int
	*continuousFire

	font                        *Font
	doubleHitFastTexture        sdl.Texture
	doubleHitFastTextureSize    sdl.Rect
	doubleHitSlowTexture        sdl.Texture
	doubleHitSlowTextureSize    sdl.Rect
	doubleHitDisableTexture     sdl.Texture
	doubleHitDisableTextureSize sdl.Rect
}

func (ch *controlHandler) Init(r sdl.Renderer) {
	var err error
	if ch.font == nil {
		if ch.font, err = OpenFont(filepath.Join(sdl.GetBasePath(), "YaHei.Consolas.1.12.ttf"), 25); err != nil {
			panic(err)
		}
	}

	ch.doubleHitFastTexture, ch.doubleHitFastTextureSize = ch.initTextures(r, "连击模式：快")
	ch.doubleHitSlowTexture, ch.doubleHitSlowTextureSize = ch.initTextures(r, "连击模式：慢")
	ch.doubleHitDisableTexture, ch.doubleHitDisableTextureSize = ch.initTextures(r, "连击模式：关闭")
}

func (ch *controlHandler) initTextures(r sdl.Renderer, text string) (sdl.Texture, sdl.Rect) {
	if surface, err := ch.font.GetTextSurface(text, sdl.Color{}); err != nil {
		panic(err)
	} else {
		if texture, err := r.CreateTextureFromSurface(surface); err != nil {
			panic(err)
		} else {
			surface.Free()
			size := getTextureSize(texture, 50, 50)
			return texture, size
		}
	}
}

func getTextureSize(t sdl.Texture, startX, startY int32) sdl.Rect {
	_, _, w, h, _ := t.Query()
	return sdl.Rect{startX, startY, w, h}
}

func (ch *controlHandler) Render(r sdl.Renderer) {
	switch ch.doubleHit {
	case 0:
		// 关闭
		r.Copy(ch.doubleHitDisableTexture, nil, &ch.doubleHitDisableTextureSize)

	case 1:
		// 快
		r.Copy(ch.doubleHitFastTexture, nil, &ch.doubleHitFastTextureSize)

	case 2:
		// 慢
		r.Copy(ch.doubleHitSlowTexture, nil, &ch.doubleHitSlowTextureSize)
	}
}

func newControlHandler(controller Controller, keyMap, ctrlKeyMap map[int]*Point) *controlHandler {
	ch := controlHandler{controller: controller}
	controller.Register(&ch)
	ch.keyState = make(map[int]*int)
	ch.ctrlKeyState = make(map[int]*int)
	ch.keyMap = keyMap
	ch.ctrlKeyMap = ctrlKeyMap
	ch.directionController.keyMap = keyMap
	ch.doubleHit = 0
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
			b, e = ch.sendMouseEvent(AMOTION_EVENT_ACTION_UP, *ch.keyState[VisionKeyCode], ch.visionCachePointer)
			fingers.Recycle(ch.keyState[VisionKeyCode])
			ch.keyState[VisionKeyCode] = nil
			if debugOpt.Info() {
				log.Println("视角控制，松开，点：", ch.visionCachePointer)
			}
		}
		return b, e

	case eventDirectionEvent:
		return true, ch.directionController.sendMouseEvent(ch.controller)

	case eventWheelEvent:
		var b bool
		var e error
		if ch.keyState[WheelKeyCode] != nil {
			b, e = ch.sendMouseEvent(AMOTION_EVENT_ACTION_UP, *ch.keyState[WheelKeyCode], ch.wheelCachePointer)
			fingers.Recycle(ch.keyState[WheelKeyCode])
			ch.keyState[WheelKeyCode] = nil
		}
		return b, e

	case sdl.MOUSEMOTION:
		return ch.handleMouseMotion(event.(*sdl.MouseMotionEvent))

	case sdl.MOUSEBUTTONDOWN:
		return ch.handleMouseButtonDown(event.(*sdl.MouseButtonEvent))

	case sdl.MOUSEBUTTONUP:
		return ch.handleMouseButtonUp(event.(*sdl.MouseButtonEvent))

	case sdl.MOUSEWHEEL:
		return ch.handleMouseWheelMotion(event.(*sdl.MouseWheelEvent))

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
		ch.visionCachePointer = Point{950, 450}
		ch.sendEventDelay(eventVisionEventUp, mouseVisionDelay)
		if debugOpt.Info() {
			log.Println("视角控制，开始，点：", ch.visionCachePointer)
		}
		return ch.sendMouseEvent(AMOTION_EVENT_ACTION_DOWN, *ch.keyState[VisionKeyCode], ch.visionCachePointer)
	} else {
		deltaX := fixMouseBlock(event.XRel)
		deltaY := fixMouseBlock(event.YRel)
		ch.visionCachePointer.X = uint16(int32(ch.visionCachePointer.X) + deltaX)
		ch.visionCachePointer.Y = uint16(int32(ch.visionCachePointer.Y) + deltaY + int32(delta))
		if ch.outside(&ch.visionCachePointer) {
			b, e := ch.sendMouseEvent(AMOTION_EVENT_ACTION_UP, *ch.keyState[VisionKeyCode], ch.visionCachePointer)
			fingers.Recycle(ch.keyState[VisionKeyCode])
			ch.keyState[VisionKeyCode] = nil
			if debugOpt.Info() {
				log.Printf("视角控制(%d, %d)，超出范围，点：%s\n", deltaX, deltaY, ch.visionCachePointer)
			}
			return b, e
		} else {
			ch.sendEventDelay(eventVisionEventUp, mouseVisionDelay)
			if debugOpt.Info() {
				log.Printf("视角控制(%d, %d)，点：%s\n", deltaX, deltaY, ch.visionCachePointer)
			}
			return ch.sendMouseEvent(AMOTION_EVENT_ACTION_MOVE, *ch.keyState[VisionKeyCode], ch.visionCachePointer)
		}
	}
}

func (ch *controlHandler) stopContinuousFire() {
	if ch.continuousFire != nil {
		ch.continuousFire.Stop()
		ch.continuousFire = nil
	}
}

func (ch *controlHandler) startContinuousFire(interval time.Duration) {
	if ch.continuousFire == nil {
		ch.continuousFire = new(continuousFire)
		ch.continuousFire.Point = *ch.keyMap[FireKeyCode]
		ch.continuousFire.Start(ch.controller, interval)
	} else {
		ch.continuousFire.SetInterval(interval)
	}
}

func (ch *controlHandler) startMainPointerMotion(x, y int32) {
	if ch.keyState[mainPointerKeyCode] == nil {
		ch.keyState[mainPointerKeyCode] = fingers.GetId()
		ch.sendMouseEvent(AMOTION_EVENT_ACTION_DOWN, *ch.keyState[mainPointerKeyCode], Point{uint16(x), uint16(y)})
	} else {
		panic("main pointer state error")
	}
}

func (ch *controlHandler) continueMainPointerMotion(x, y int32) {
	if ch.keyState[mainPointerKeyCode] != nil {
		ch.sendMouseEvent(AMOTION_EVENT_ACTION_MOVE, *ch.keyState[mainPointerKeyCode], Point{uint16(x), uint16(y)})
	} else {
		panic("main pointer state error")
	}
}

func (ch *controlHandler) stopMainPointerMotion(x, y int32) {
	if ch.keyState[mainPointerKeyCode] != nil {
		ch.sendMouseEvent(AMOTION_EVENT_ACTION_UP, *ch.keyState[mainPointerKeyCode], Point{uint16(x), uint16(y)})
		fingers.Recycle(ch.keyState[mainPointerKeyCode])
		ch.keyState[mainPointerKeyCode] = nil
	}
}

func (ch *controlHandler) handleMouseMotion(event *sdl.MouseMotionEvent) (bool, error) {
	if sdl.GetRelativeMouseMode() {
		// 无论如何，停止 mainPointer 的事件分发
		ch.stopMainPointerMotion(event.X, event.Y)
		return ch.visionMoving(event, 0)
	} else {
		// 无论如何，关闭连击宏操作
		ch.stopContinuousFire()

		if ch.keyState[VisionKeyCode] != nil {
			b, e := ch.sendMouseEvent(AMOTION_EVENT_ACTION_UP, *ch.keyState[VisionKeyCode], ch.visionCachePointer)
			fingers.Recycle(ch.keyState[VisionKeyCode])
			ch.keyState[VisionKeyCode] = nil
			return b, e
		}

		if event.State == sdl.BUTTON_LEFT {
			ch.continueMainPointerMotion(event.X, event.Y)
		}
	}

	return true, nil
}

func (ch *controlHandler) handleMouseButtonDown(event *sdl.MouseButtonEvent) (bool, error) {
	// 鼠标左键
	if event.Button == sdl.BUTTON_LEFT {
		// 无论如何，关闭连击宏操作
		ch.stopContinuousFire()

		if sdl.GetRelativeMouseMode() {
			// 无论如何，停止 mainPointer 的事件分发
			ch.stopMainPointerMotion(event.X, event.Y)

			switch ch.doubleHit {
			case 0:
				if ch.keyState[FireKeyCode] == nil {
					ch.keyState[FireKeyCode] = fingers.GetId()
					if debugOpt.Debug() {
						log.Println("按下开火键")
					}
					ch.sendMouseEvent(AMOTION_EVENT_ACTION_DOWN, *ch.keyState[FireKeyCode], *ch.keyMap[FireKeyCode])
				}
				if debugOpt.Debug() {
					log.Println("正常开火")
				}

			case 1:
				ch.startContinuousFire(60 * time.Millisecond)
				if debugOpt.Debug() {
					log.Println("连击快速")
				}

			case 2:
				ch.startContinuousFire(250 * time.Millisecond)
				if debugOpt.Debug() {
					log.Println("连击慢速")
				}
			}
		} else {
			ch.startMainPointerMotion(event.X, event.Y)
		}
	}

	return true, nil
}

func (ch *controlHandler) handleMouseButtonUp(event *sdl.MouseButtonEvent) (bool, error) {
	// 鼠标左键
	if event.Button == sdl.BUTTON_LEFT {
		// 无论如何，关闭连击宏操作
		ch.stopContinuousFire()
		// 无论如何，停止 mainPointer 的事件分发
		ch.stopMainPointerMotion(event.X, event.Y)

		if sdl.GetRelativeMouseMode() {
			if ch.keyState[FireKeyCode] != nil {
				b, e := ch.sendMouseEvent(AMOTION_EVENT_ACTION_UP, *ch.keyState[FireKeyCode], *ch.keyMap[FireKeyCode])
				fingers.Recycle(ch.keyState[FireKeyCode])
				ch.keyState[FireKeyCode] = nil
				if debugOpt.Debug() {
					log.Println("松开开火键")
				}
				return b, e
			}
		}
	} else if event.Button == sdl.BUTTON_RIGHT {
		ch.doubleHit = (ch.doubleHit + 1) % 3
		if debugOpt.Debug() {
			log.Printf("连击模式:%d\n", ch.doubleHit)
		}
	}

	return true, nil
}

func (ch *controlHandler) handleKeyDown(event *sdl.KeyboardEvent) (bool, error) {
	alt := event.Keysym.Mod&(sdl.KMOD_RALT|sdl.KMOD_LALT) != 0
	if alt {
		return true, nil
	}
	ctrl := event.Keysym.Mod&(sdl.KMOD_RCTRL|sdl.KMOD_LCTRL) != 0
	if ctrl {
		keyCode := int(event.Keysym.Sym)
		if poi := ch.ctrlKeyMap[keyCode]; poi != nil {
			if ch.ctrlKeyState[keyCode] == nil {
				ch.ctrlKeyState[keyCode] = fingers.GetId()
				return ch.sendMouseEvent(AMOTION_EVENT_ACTION_DOWN, *ch.ctrlKeyState[keyCode], *poi)
			} else {
				return ch.sendMouseEvent(AMOTION_EVENT_ACTION_MOVE, *ch.ctrlKeyState[keyCode], *poi)
			}
		}
	} else {
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
				ch.directionController.Start()
				return true, nil

			case sdl.K_s:
				ch.directionController.backDown()
				ch.directionController.Start()
				return true, nil

			case sdl.K_a:
				ch.directionController.leftDown()
				ch.directionController.Start()
				return true, nil

			case sdl.K_d:
				ch.directionController.rightDown()
				ch.directionController.Start()
				return true, nil
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
		keyCode := int(event.Keysym.Sym)
		if poi := ch.ctrlKeyMap[keyCode]; poi != nil {
			b, e := ch.sendMouseEvent(AMOTION_EVENT_ACTION_UP, *ch.ctrlKeyState[keyCode], *poi)
			fingers.Recycle(ch.ctrlKeyState[keyCode])
			ch.ctrlKeyState[keyCode] = nil
			return b, e
		} else {
			switch event.Keysym.Sym {
			case sdl.K_x:
				sdl.SetRelativeMouseMode(!sdl.GetRelativeMouseMode())
			case sdl.K_z:
				ch.doubleHit = 0
				if debugOpt.Debug() {
					log.Printf("连击模式:%d\n", ch.doubleHit)
				}
			}
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

			if ch.keyState[keyCode] != nil {
				b, e := ch.sendMouseEvent(AMOTION_EVENT_ACTION_UP, *ch.keyState[keyCode], *poi)
				fingers.Recycle(ch.keyState[keyCode])
				ch.keyState[keyCode] = nil
				return b, e
			}
		} else {
			switch event.Keysym.Sym {
			case sdl.K_w:
				ch.directionController.frontUp()
				return true, nil

			case sdl.K_s:
				ch.directionController.backUp()
				return true, nil

			case sdl.K_a:
				ch.directionController.leftUp()
				return true, nil

			case sdl.K_d:
				ch.directionController.rightUp()
				return true, nil

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

func (ch *controlHandler) handleMouseWheelMotion(event *sdl.MouseWheelEvent) (bool, error) {
	log.Printf("x: %d, y: %d, direction: %d\n", event.X, event.Y, event.Direction)
	if ch.keyState[WheelKeyCode] == nil {
		ch.keyState[WheelKeyCode] = fingers.GetId()
		ch.wheelCachePointer = *ch.keyMap[sdl.K_g]
		ch.sendEventDelay(eventWheelEvent, 150*time.Millisecond)
		return ch.sendMouseEvent(AMOTION_EVENT_ACTION_DOWN, *ch.keyState[WheelKeyCode], ch.wheelCachePointer)
	} else {
		deltaY := event.Y * 20
		tmp := int32(ch.wheelCachePointer.Y) + deltaY
		if tmp < 0 {
			tmp = 0
		} else if tmp > 800 {
			tmp = 800
		}
		ch.wheelCachePointer.Y = uint16(tmp)
		ch.sendEventDelay(eventWheelEvent, 150*time.Millisecond)
		return ch.sendMouseEvent(AMOTION_EVENT_ACTION_MOVE, *ch.keyState[WheelKeyCode], ch.wheelCachePointer)
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
