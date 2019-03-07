package scrcpy

import (
	"bytes"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/ClarkGuan/go-sdl2/sdl"
)

const (
	mainPointerKeyCode = 500 + iota
	FireKeyCode
	VisionBoundTopLeft
	VisionBoundBottomRight
	FrontKeyCode
	BackKeyCode
	wheelKeyCode
)

const eventDirectionEvent = sdl.USEREVENT + 4
const eventWheelEvent = sdl.USEREVENT + 5

var mouseIntervalArray = []time.Duration{
	0,
	30 * time.Millisecond,
}

var gunPressArray = []*GunPressConfig{
	nil,
	{3, 28 * time.Millisecond},
}

func setConfigs(hit []time.Duration, stables []*GunPressConfig) {
	if len(hit) > 0 {
		mouseIntervalArray = mouseIntervalArray[:1]
		mouseIntervalArray = append(mouseIntervalArray, hit...)
	}
	if len(stables) > 0 {
		gunPressArray = gunPressArray[:1]
		gunPressArray = append(gunPressArray, stables...)
	}
}

type controlHandler struct {
	controller       Controller
	visionController *visionController
	set              mouseEventSet

	keyState map[int]*int
	keyMap   map[int]UserOperation

	ctrlKeyState map[int]*int
	ctrlKeyMap   map[int]UserOperation

	mouseKeyState map[uint8]*int
	mouseKeyMap   map[uint8]UserOperation

	wheelCachePointer Point

	// 自动压枪处理
	gunPress    int
	gunPressOpr *gunPressOperation

	directionController directionController
	timer               map[uint32]*time.Timer
	doubleHit           int
	*continuousFire

	font            *Font
	textTexture     *TextTexture
	displayPosition sdl.Rect
	textBuf         bytes.Buffer
}

func (ch *controlHandler) Init(r sdl.Renderer) {
	var err error
	if ch.font == nil {
		if ch.font, err = OpenFont(filepath.Join(sdl.GetBasePath(), "res", "YaHei.Consolas.1.12.ttf"), 35); err != nil {
			panic(err)
		}
	}

	ch.textTexture = new(TextTexture)
	ch.displayPosition.X = 50
	ch.displayPosition.Y = 50
}

func (ch *controlHandler) Render(r sdl.Renderer) {
	ch.textBuf.Reset()

	switch ch.doubleHit {
	case 0:
		// ignore

	default:
		fmt.Fprintf(&ch.textBuf, "连击模式：%v  ", mouseIntervalArray[ch.doubleHit%len(mouseIntervalArray)])
	}

	switch ch.gunPress {
	case 0:
		// ignore

	default:
		fmt.Fprintf(&ch.textBuf, "自动压枪：%v", gunPressArray[ch.gunPress%len(gunPressArray)])
	}

	ch.textTexture.Update(r, ch.font, ch.textBuf.String(), sdl.Color{}, &ch.displayPosition)
	ch.textTexture.Render(r, &ch.displayPosition)
}

func newControlHandler(controller Controller,
	keyMap, ctrlKeyMap map[int]UserOperation,
	mouseKeyMap map[uint8]UserOperation) *controlHandler {

	ch := controlHandler{controller: controller}
	controller.Register(&ch)
	ch.keyState = make(map[int]*int)
	ch.ctrlKeyState = make(map[int]*int)
	ch.mouseKeyState = make(map[uint8]*int)
	ch.keyMap = keyMap
	ch.ctrlKeyMap = ctrlKeyMap
	ch.mouseKeyMap = mouseKeyMap
	ch.directionController.keyMap = keyMap
	// 默认是正常模式
	ch.doubleHit = 0
	// 默认关闭自动压枪
	ch.gunPress = 0

	// 视角控制
	ch.visionController = newVisionController(controller,
		keyMap[VisionBoundTopLeft].(*Point),
		keyMap[VisionBoundBottomRight].(*Point))
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

	// 处理视角 SDL 事件
	if ch.visionController.handleSdlEvent(event.GetType()) {
		return true, nil
	}

	switch event.GetType() {
	case eventDirectionEvent:
		return true, ch.directionController.sendMouseEvent(ch.controller)

	case eventWheelEvent:
		var b bool
		var e error
		if ch.keyState[wheelKeyCode] != nil {
			b, e = ch.sendMouseEvent(AMOTION_EVENT_ACTION_UP, *ch.keyState[wheelKeyCode], ch.wheelCachePointer)
			fingers.Recycle(ch.keyState[wheelKeyCode])
			ch.keyState[wheelKeyCode] = nil
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

func (ch *controlHandler) stopContinuousFire() {
	if ch.continuousFire != nil {
		ch.continuousFire.Stop()
		ch.continuousFire = nil
	}
}

func (ch *controlHandler) startContinuousFire(interval time.Duration) {
	if ch.continuousFire == nil {
		ch.continuousFire = new(continuousFire)
		ch.continuousFire.Point = *(ch.keyMap[FireKeyCode].(*Point))
		ch.continuousFire.Start(ch.controller, interval)
	} else {
		ch.continuousFire.SetInterval(interval)
	}
}

func (ch *controlHandler) stopGunPress() {
	if ch.gunPressOpr != nil {
		ch.gunPressOpr.Stop()
		ch.gunPressOpr = nil
	}
}

func (ch *controlHandler) startGunPress(interval time.Duration, delta int) {
	if ch.gunPress > 0 {
		if ch.gunPressOpr == nil {
			ch.gunPressOpr = new(gunPressOperation)
			ch.gunPressOpr.Start(ch.visionController, *gunPressArray[ch.gunPress%len(gunPressArray)])
		} else {
			ch.gunPressOpr.SetValues(*gunPressArray[ch.gunPress%len(gunPressArray)])
		}
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
		ch.visionController.visionControl(event.XRel, event.YRel)
		return true, nil
	} else {
		// 无论如何，关闭连击宏操作
		ch.stopContinuousFire()

		// 视角控制手势退出
		ch.visionController.fingerUp()

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
		// 无论如何，关闭压枪操作
		ch.stopGunPress()

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
					ch.sendMouseEvent(AMOTION_EVENT_ACTION_DOWN, *ch.keyState[FireKeyCode], *(ch.keyMap[FireKeyCode].(*Point)))
				}
				if debugOpt.Debug() {
					log.Println("正常开火")
				}

			default:
				ch.startContinuousFire(mouseIntervalArray[ch.doubleHit])
			}

			ch.startGunPress(30*time.Millisecond, 1)
		} else {
			ch.startMainPointerMotion(event.X, event.Y)
		}
	} else if ch.mouseKeyMap[event.Button] != nil {
		if p, ok := ch.mouseKeyMap[event.Button].(*Point); ok {
			if ch.mouseKeyState[event.Button] == nil {
				ch.mouseKeyState[event.Button] = fingers.GetId()
				ch.sendMouseEvent(AMOTION_EVENT_ACTION_DOWN, *ch.mouseKeyState[event.Button], *p)
			}
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
		// 无论如何，关闭压枪操作
		ch.stopGunPress()

		if sdl.GetRelativeMouseMode() {
			if ch.keyState[FireKeyCode] != nil {
				b, e := ch.sendMouseEvent(AMOTION_EVENT_ACTION_UP, *ch.keyState[FireKeyCode], *(ch.keyMap[FireKeyCode].(*Point)))
				fingers.Recycle(ch.keyState[FireKeyCode])
				ch.keyState[FireKeyCode] = nil
				if debugOpt.Debug() {
					log.Println("松开开火键")
				}
				return b, e
			}
		}
	} else if ch.mouseKeyMap[event.Button] != nil {
		if p, ok := ch.mouseKeyMap[event.Button].(*Point); ok {
			if ch.mouseKeyState[event.Button] != nil {
				ch.sendMouseEvent(AMOTION_EVENT_ACTION_UP, *ch.mouseKeyState[event.Button], *p)
				fingers.Recycle(ch.mouseKeyState[event.Button])
				ch.mouseKeyState[event.Button] = nil
			}
		} else if pms, ok := ch.mouseKeyMap[event.Button].([]*PointMacro); ok {
			ca := newControllerAnimation(ch.controller, pms)
			ca.start()
		}
	}

	return true, nil
}

func (ch *controlHandler) handleKeyDown(event *sdl.KeyboardEvent) (bool, error) {
	if event.Repeat > 0 {
		// 减少事件传递，提升效率，降低传输数据量
		return true, nil
	}

	alt := event.Keysym.Mod&(sdl.KMOD_RALT|sdl.KMOD_LALT) != 0
	if alt {
		return true, nil
	}
	ctrl := event.Keysym.Mod&(sdl.KMOD_RCTRL|sdl.KMOD_LCTRL) != 0
	if ctrl {
		switch event.Keysym.Sym {
		case sdl.K_h:
			kce := keyCodeEvent{
				action:  AKEY_EVENT_ACTION_DOWN,
				keyCode: AKEYCODE_HOME,
			}
			ch.controller.PushEvent(&kce)
			return true, nil

		case sdl.K_b:
			kce := keyCodeEvent{
				action:  AKEY_EVENT_ACTION_DOWN,
				keyCode: AKEYCODE_BACK,
			}
			ch.controller.PushEvent(&kce)
			return true, nil

		case sdl.K_s:
			kce := keyCodeEvent{
				action:  AKEY_EVENT_ACTION_DOWN,
				keyCode: AKEYCODE_APP_SWITCH,
			}
			ch.controller.PushEvent(&kce)
			return true, nil

		case sdl.K_p:
			kce := keyCodeEvent{
				action:  AKEY_EVENT_ACTION_DOWN,
				keyCode: AKEYCODE_POWER,
			}
			ch.controller.PushEvent(&kce)
			return true, nil

		case sdl.K_m:
			kce := keyCodeEvent{
				action:  AKEY_EVENT_ACTION_DOWN,
				keyCode: AKEYCODE_MENU,
			}
			ch.controller.PushEvent(&kce)
			return true, nil

		case sdl.K_SEMICOLON:
			kce := keyCodeEvent{
				action:  AKEY_EVENT_ACTION_DOWN,
				keyCode: AKEYCODE_VOLUME_UP,
			}
			ch.controller.PushEvent(&kce)
			return true, nil

		case sdl.K_QUOTE:
			kce := keyCodeEvent{
				action:  AKEY_EVENT_ACTION_DOWN,
				keyCode: AKEYCODE_VOLUME_DOWN,
			}
			ch.controller.PushEvent(&kce)
			return true, nil
		}

		keyCode := int(event.Keysym.Sym)
		if ch.ctrlKeyMap[keyCode] != nil {
			if p, ok := ch.ctrlKeyMap[keyCode].(*Point); ok {
				if ch.ctrlKeyState[keyCode] == nil {
					ch.ctrlKeyState[keyCode] = fingers.GetId()
					return ch.sendMouseEvent(AMOTION_EVENT_ACTION_DOWN, *ch.ctrlKeyState[keyCode], *p)
				} else {
					return ch.sendMouseEvent(AMOTION_EVENT_ACTION_MOVE, *ch.ctrlKeyState[keyCode], *p)
				}
			}
		}
	} else {
		// w,s,a,d 四个按键不能被自定义按键覆盖！
		switch event.Keysym.Sym {
		case sdl.K_w:
			fallthrough
		case sdl.K_UP:
			ch.directionController.frontDown()
			ch.directionController.Start()
			return true, nil

		case sdl.K_s:
			fallthrough
		case sdl.K_DOWN:
			ch.directionController.backDown()
			ch.directionController.Start()
			return true, nil

		case sdl.K_a:
			fallthrough
		case sdl.K_LEFT:
			ch.directionController.leftDown()
			ch.directionController.Start()
			return true, nil

		case sdl.K_d:
			fallthrough
		case sdl.K_RIGHT:
			ch.directionController.rightDown()
			ch.directionController.Start()
			return true, nil
		}

		keyCode := int(event.Keysym.Sym)
		if ch.keyMap[keyCode] != nil {
			if p, ok := ch.keyMap[keyCode].(*Point); ok {
				if ch.keyState[keyCode] == nil {
					ch.keyState[keyCode] = fingers.GetId()
					return ch.sendMouseEvent(AMOTION_EVENT_ACTION_DOWN, *ch.keyState[keyCode], *p)
				} else {
					return ch.sendMouseEvent(AMOTION_EVENT_ACTION_MOVE, *ch.keyState[keyCode], *p)
				}
			} else if sp, ok := ch.keyMap[keyCode].(*SPoint); ok {
				if ch.keyState[keyCode] == nil {
					ch.keyState[keyCode] = fingers.GetId()
					return ch.sendMouseEvent(AMOTION_EVENT_ACTION_DOWN, *ch.keyState[keyCode], Point(*sp))
				} else {
					return ch.sendMouseEvent(AMOTION_EVENT_ACTION_MOVE, *ch.keyState[keyCode], Point(*sp))
				}
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
		// ctrl+x, ctrl+0, ctrl+-, ctrl+= 按键不允许自定义按键覆盖
		switch event.Keysym.Sym {
		case sdl.K_h:
			kce := keyCodeEvent{
				action:  AKEY_EVENT_ACTION_UP,
				keyCode: AKEYCODE_HOME,
			}
			ch.controller.PushEvent(&kce)
			return true, nil

		case sdl.K_b:
			kce := keyCodeEvent{
				action:  AKEY_EVENT_ACTION_UP,
				keyCode: AKEYCODE_BACK,
			}
			ch.controller.PushEvent(&kce)
			return true, nil

		case sdl.K_s:
			kce := keyCodeEvent{
				action:  AKEY_EVENT_ACTION_UP,
				keyCode: AKEYCODE_APP_SWITCH,
			}
			ch.controller.PushEvent(&kce)
			return true, nil

		case sdl.K_p:
			kce := keyCodeEvent{
				action:  AKEY_EVENT_ACTION_UP,
				keyCode: AKEYCODE_POWER,
			}
			ch.controller.PushEvent(&kce)
			return true, nil

		case sdl.K_m:
			kce := keyCodeEvent{
				action:  AKEY_EVENT_ACTION_UP,
				keyCode: AKEYCODE_MENU,
			}
			ch.controller.PushEvent(&kce)
			return true, nil

		case sdl.K_SEMICOLON:
			kce := keyCodeEvent{
				action:  AKEY_EVENT_ACTION_UP,
				keyCode: AKEYCODE_VOLUME_UP,
			}
			ch.controller.PushEvent(&kce)
			return true, nil

		case sdl.K_QUOTE:
			kce := keyCodeEvent{
				action:  AKEY_EVENT_ACTION_UP,
				keyCode: AKEYCODE_VOLUME_DOWN,
			}
			ch.controller.PushEvent(&kce)
			return true, nil

		case sdl.K_x:
			sdl.SetRelativeMouseMode(!sdl.GetRelativeMouseMode())
			return true, nil

		case sdl.K_0:
			ch.doubleHit = 0
			ch.gunPress = 0
			return true, nil
		}

		keyCode := int(event.Keysym.Sym)
		if ch.ctrlKeyMap[keyCode] != nil {
			if p, ok := ch.ctrlKeyMap[keyCode].(*Point); ok {
				ch.sendMouseEvent(AMOTION_EVENT_ACTION_UP, *ch.ctrlKeyState[keyCode], *p)
				fingers.Recycle(ch.ctrlKeyState[keyCode])
				ch.ctrlKeyState[keyCode] = nil
				return true, nil
			} else if pms, ok := ch.ctrlKeyMap[keyCode].([]*PointMacro); ok {
				ca := newControllerAnimation(ch.controller, pms)
				ca.start()
				return true, nil
			}
		}
	} else {
		// w,s,a,d 按键的方案不能被自定义按键方案覆盖
		switch event.Keysym.Sym {
		case sdl.K_F2:
			ch.gunPress = (ch.gunPress + 1) % len(gunPressArray)
			return true, nil

		case sdl.K_F1:
			ch.doubleHit = (ch.doubleHit + 1) % len(mouseIntervalArray)
			return true, nil

		case sdl.K_w:
			fallthrough
		case sdl.K_UP:
			ch.directionController.frontUp()
			return true, nil

		case sdl.K_s:
			fallthrough
		case sdl.K_DOWN:
			ch.directionController.backUp()
			return true, nil

		case sdl.K_a:
			fallthrough
		case sdl.K_LEFT:
			ch.directionController.leftUp()
			return true, nil

		case sdl.K_d:
			fallthrough
		case sdl.K_RIGHT:
			ch.directionController.rightUp()
			return true, nil
		}

		keyCode := int(event.Keysym.Sym)
		if ch.keyMap[keyCode] != nil {
			if p, ok := ch.keyMap[keyCode].(*Point); ok {
				if ch.keyState[keyCode] != nil {
					b, e := ch.sendMouseEvent(AMOTION_EVENT_ACTION_UP, *ch.keyState[keyCode], *p)
					fingers.Recycle(ch.keyState[keyCode])
					ch.keyState[keyCode] = nil
					return b, e
				}
			} else if sp, ok := ch.keyMap[keyCode].(*SPoint); ok {
				sdl.SetRelativeMouseMode(!sdl.GetRelativeMouseMode())
				if ch.keyState[keyCode] != nil {
					b, e := ch.sendMouseEvent(AMOTION_EVENT_ACTION_UP, *ch.keyState[keyCode], Point(*sp))
					fingers.Recycle(ch.keyState[keyCode])
					ch.keyState[keyCode] = nil
					return b, e
				}
			} else if pms, ok := ch.keyMap[keyCode].([]*PointMacro); ok {
				ca := newControllerAnimation(ch.controller, pms)
				ca.start()
			}
		}
	}

	return true, nil
}

func (ch *controlHandler) handleMouseWheelMotion(event *sdl.MouseWheelEvent) (bool, error) {
	if debugOpt.Debug() {
		log.Printf("x: %d, y: %d, direction: %d\n", event.X, event.Y, event.Direction)
	}
	if ch.keyState[wheelKeyCode] == nil {
		ch.keyState[wheelKeyCode] = fingers.GetId()
		ch.wheelCachePointer = *(ch.keyMap[sdl.K_g].(*Point))
		ch.sendEventDelay(eventWheelEvent, 150*time.Millisecond)
		return ch.sendMouseEvent(AMOTION_EVENT_ACTION_DOWN, *ch.keyState[wheelKeyCode], ch.wheelCachePointer)
	} else {
		deltaY := event.Y * 10
		tmp := int32(ch.wheelCachePointer.Y) + deltaY
		if tmp < 0 {
			tmp = 0
		} else if tmp > 800 {
			tmp = 800
		}
		ch.wheelCachePointer.Y = uint16(tmp)
		ch.sendEventDelay(eventWheelEvent, 150*time.Millisecond)
		return ch.sendMouseEvent(AMOTION_EVENT_ACTION_MOVE, *ch.keyState[wheelKeyCode], ch.wheelCachePointer)
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

func (ch *controlHandler) stopEvent(typ uint32) {
	if ch.timer == nil {
		ch.timer = make(map[uint32]*time.Timer)
	}

	if ch.timer[typ] != nil {
		ch.timer[typ].Stop()
	}
}
