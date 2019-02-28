package scrcpy

import (
	"log"
	"time"

	"github.com/ClarkGuan/go-sdl2/sdl"
)

const DefaultMouseSensitive = .085

// 鼠标精度控制
var mouseSensitive = DefaultMouseSensitive

// 自动释放手势时间间隔
const mouseVisionDelay = time.Millisecond * 500

// 注册到 SDL 中的自定义事件
const eventVisionEventUp = sdl.USEREVENT + 3

// 视野控制器
type visionController struct {
	controller  Controller
	topLeft     Point
	bottomRight Point

	center     *Point
	cachePoint Point
	id         *int
	timer      *time.Timer
}

func newVisionController(controller Controller, topLeft, bottomRight *Point) *visionController {
	return &visionController{
		controller:  controller,
		topLeft:     *topLeft,
		bottomRight: *bottomRight,
	}
}

func (v *visionController) outside(p *Point) bool {
	ret := false
	minW := uint16(v.topLeft.X)
	maxW := uint16(v.bottomRight.X)
	if p.X < minW {
		ret = true
		p.X = minW
	} else if p.X > maxW {
		ret = true
		p.X = maxW
	}

	minH := uint16(v.topLeft.Y)
	maxH := uint16(v.bottomRight.Y)
	if p.Y < minH {
		ret = true
		p.Y = minH
	} else if p.Y > maxH {
		ret = true
		p.Y = maxH
	}

	return ret
}

func (v *visionController) getVisionCenterPoint() *Point {
	if v.center == nil {
		v.center = &Point{
			X: (v.topLeft.X + v.bottomRight.X) >> 1,
			Y: (v.topLeft.Y + v.bottomRight.Y) >> 1,
		}
	}
	return v.center
}

func fixMouseBlock(x int32) int32 {
	fx := float64(x)
	ret := int32(fx*mouseSensitive + .5)
	if ret == 0 && x != 0 {
		if x > 0 {
			ret = 1
		} else {
			ret = -1
		}
	}
	return ret
}

func (v *visionController) sendMouseEvent(action androidMotionEventAction, id int, p Point) error {
	sme := singleMouseEvent{action: action}
	sme.id = id
	sme.Point = p
	return v.controller.PushEvent(&sme)
}

func (v *visionController) sendEventDelay(duration time.Duration) {
	if v.timer != nil {
		v.timer.Reset(duration)
	} else {
		v.timer = time.AfterFunc(duration, func() {
			sdl.PushEvent(&sdl.UserEvent{Type: eventVisionEventUp})
		})
	}
}

func (v *visionController) stopEventDelay() {
	if v.timer != nil {
		v.timer.Stop()
	}
}

// 处理 SDL 事件队列中收到的事件
func (v *visionController) handleSdlEvent(typ uint32) bool {
	if typ == eventVisionEventUp {
		v.fingerUp()
		return true
	} else {
		return false
	}
}

func (v *visionController) fingerUp() {
	v.stopEventDelay()
	if v.id != nil {
		v.sendMouseEvent(AMOTION_EVENT_ACTION_UP, *v.id, v.cachePoint)
		fingers.Recycle(v.id)
		v.id = nil
		if debugOpt.Info() {
			log.Println("视角控制，松开，点：", v.cachePoint)
		}
	}
}

func (v *visionController) fingerDown() {
	if v.id == nil {
		v.id = fingers.GetId()
		v.cachePoint = *v.getVisionCenterPoint()
		v.sendEventDelay(mouseVisionDelay)
		if debugOpt.Info() {
			log.Println("视角控制，开始，点：", v.cachePoint)
		}
		v.sendMouseEvent(AMOTION_EVENT_ACTION_DOWN, *v.id, v.cachePoint)
	}
}

func (v *visionController) fingerMove(x, y int32, accurate bool) {
	if !accurate {
		x = fixMouseBlock(x)
		y = fixMouseBlock(y)
	}
	v.cachePoint.X = uint16(int32(v.cachePoint.X) + x)
	v.cachePoint.Y = uint16(int32(v.cachePoint.Y) + y)
	if v.outside(&v.cachePoint) {
		v.fingerUp()
	} else {
		v.sendEventDelay(mouseVisionDelay)
		if debugOpt.Info() {
			log.Printf("视角控制(%d, %d)，点：%s\n", x, y, v.cachePoint)
		}
		v.sendMouseEvent(AMOTION_EVENT_ACTION_MOVE, *v.id, v.cachePoint)
	}
}

func (v *visionController) visionControl(x, y int32) {
	if v.id == nil {
		v.fingerDown()
	} else {
		v.fingerMove(x, y, false)
	}
}

func (v *visionController) visionControl2(x, y int32) {
	if v.id == nil {
		v.fingerDown()
	} else {
		v.fingerMove(x, y, true)
	}
}
