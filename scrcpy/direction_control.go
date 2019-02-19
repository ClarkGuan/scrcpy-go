package scrcpy

import (
	"sync/atomic"
	"time"

	"github.com/ClarkGuan/go-sdl2/sdl"
)

type Direction int

const (
	frontDirection Direction = 1 << iota
	backDirection
	leftDirection
	rightDirection
)

// TODO 小米 8：斜上 80，正上 150
// TODO 三星 s6 edge+：斜上 90，正上 185
const deltaDirectionObliqueMovement = 90
const deltaDirectionRightUpMovement = 185

type directionController struct {
	direction   Direction
	cachePoint  Point
	middlePoint *Point
	radius      uint16
	keyMap      map[int]UserOperation
	id          *int
	startFlag   int32
	animator
}

func (dc *directionController) frontDown() {
	dc.direction |= frontDirection
}

func (dc *directionController) frontUp() {
	dc.direction &^= frontDirection
}

func (dc *directionController) isFrontDown() bool {
	return dc.direction&frontDirection != 0
}

func (dc *directionController) backDown() {
	dc.direction |= backDirection
}

func (dc *directionController) backUp() {
	dc.direction &^= backDirection
}

func (dc *directionController) isBackDown() bool {
	return dc.direction&backDirection != 0
}

func (dc *directionController) leftDown() {
	dc.direction |= leftDirection
}

func (dc *directionController) leftUp() {
	dc.direction &^= leftDirection
}

func (dc *directionController) isLeftDown() bool {
	return dc.direction&leftDirection != 0
}

func (dc *directionController) rightDown() {
	dc.direction |= rightDirection
}

func (dc *directionController) rightUp() {
	dc.direction &^= rightDirection
}

func (dc *directionController) isRightDown() bool {
	return dc.direction&rightDirection != 0
}

func (dc *directionController) allUp() bool {
	return dc.direction == 0
}

func (dc *directionController) prepare() {
	if dc.middlePoint == nil {
		dc.middlePoint = new(Point)
		frontP, backP := dc.keyMap[FrontKeyCode].(*Point), dc.keyMap[BackKeyCode].(*Point)
		dc.middlePoint.X = frontP.X
		dc.middlePoint.Y = (frontP.Y + backP.Y) >> 1
		dc.radius = dc.middlePoint.Y - frontP.Y
	}
}

func (dc *directionController) getPoint(repeat bool) *Point {
	dc.prepare()
	dc.cachePoint = *dc.middlePoint

	if dc.isFrontDown() {
		dc.cachePoint.Y -= dc.radius
	}

	if dc.isLeftDown() {
		dc.cachePoint.X -= dc.radius
	}

	if dc.isRightDown() {
		dc.cachePoint.X += dc.radius
	}

	if dc.isBackDown() {
		dc.cachePoint.Y += dc.radius
	}

	if repeat {
		if dc.cachePoint.Y < dc.middlePoint.Y {
			if dc.cachePoint.X < dc.middlePoint.X {
				dc.cachePoint.Y -= deltaDirectionObliqueMovement
				dc.cachePoint.X -= deltaDirectionObliqueMovement
			} else if dc.cachePoint.X > dc.middlePoint.X {
				dc.cachePoint.Y -= deltaDirectionObliqueMovement
				dc.cachePoint.X += deltaDirectionObliqueMovement
			} else {
				dc.cachePoint.Y -= deltaDirectionRightUpMovement
			}
		}
	}

	return &dc.cachePoint
}

func (dc *directionController) Start() {
	for {
		if atomic.LoadInt32(&dc.startFlag) == 1 {
			return
		}
		if atomic.CompareAndSwapInt32(&dc.startFlag, 0, 1) {
			dc.animator.InProgress = dc.inProgress
			dc.animator.Start(nil)
			return
		}
	}
}

func (dc *directionController) inProgress(data interface{}) time.Duration {
	if atomic.LoadInt32(&dc.startFlag) == 0 {
		return 0
	} else {
		sdl.PushEvent(&sdl.UserEvent{Type: eventDirectionEvent})
		return time.Millisecond * 80
	}
}

func (dc *directionController) sendMouseEvent(controller Controller) error {
	if dc.id == nil {
		if dc.allUp() {
			atomic.StoreInt32(&dc.startFlag, 0)
			return nil
		}

		dc.id = fingers.GetId()
		point := dc.getPoint(false)
		sme := singleMouseEvent{action: AMOTION_EVENT_ACTION_DOWN}
		sme.id = *dc.id
		sme.Point = *point
		return controller.PushEvent(&sme)
	} else {
		point := &dc.cachePoint
		sme := singleMouseEvent{}
		if dc.allUp() {
			sme.action = AMOTION_EVENT_ACTION_UP
		} else {
			sme.action = AMOTION_EVENT_ACTION_MOVE
			point = dc.getPoint(true)
		}
		sme.id = *dc.id
		sme.Point = *point
		b := controller.PushEvent(&sme)
		if dc.allUp() {
			fingers.Recycle(dc.id)
			dc.id = nil
			atomic.StoreInt32(&dc.startFlag, 0)
		}
		return b
	}
}
