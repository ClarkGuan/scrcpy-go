package scrcpy

type Direction int

const maxDirectionLen = 50

const (
	frontDirection Direction = 1 << iota
	backDirection
	leftDirection
	rightDirection
)

type directionController struct {
	direction  Direction
	cachePoint Point
	keyMap     map[int]*Point
	id         *int
}

func (dc *directionController) frontDown() {
	if !dc.isBackDown() {
		dc.direction |= frontDirection
	}
}

func (dc *directionController) frontUp() {
	dc.direction &^= frontDirection
}

func (dc *directionController) isFrontDown() bool {
	return dc.direction&frontDirection != 0
}

func (dc *directionController) backDown() {
	if !dc.isFrontDown() {
		dc.direction |= backDirection
	}
}

func (dc *directionController) backUp() {
	dc.direction &^= backDirection
}

func (dc *directionController) isBackDown() bool {
	return dc.direction&backDirection != 0
}

func (dc *directionController) leftDown() {
	if !dc.isRightDown() {
		dc.direction |= leftDirection
	}
}

func (dc *directionController) leftUp() {
	dc.direction &^= leftDirection
}

func (dc *directionController) isLeftDown() bool {
	return dc.direction&leftDirection != 0
}

func (dc *directionController) rightDown() {
	if !dc.isLeftDown() {
		dc.direction |= rightDirection
	}
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

func (dc *directionController) getPoint(repeat bool) *Point {
	if dc.isFrontDown() {
		dc.cachePoint = *dc.keyMap[FrontKeyCode]
		if repeat {
			dc.cachePoint.Y -= maxDirectionLen
		}
		if dc.isLeftDown() {
			dc.cachePoint.X = dc.keyMap[LeftKeyCode].X
			if repeat {
				dc.cachePoint.X -= maxDirectionLen
			}
		} else if dc.isRightDown() {
			dc.cachePoint.X = dc.keyMap[RightKeyCode].X
			if repeat {
				dc.cachePoint.X += maxDirectionLen
			}
		}
	} else if dc.isBackDown() {
		dc.cachePoint = *dc.keyMap[BackKeyCode]
		if dc.isLeftDown() {
			dc.cachePoint.X = dc.keyMap[LeftKeyCode].X
		} else if dc.isRightDown() {
			dc.cachePoint.X = dc.keyMap[RightKeyCode].X
		}
	} else if dc.isLeftDown() {
		dc.cachePoint = *dc.keyMap[LeftKeyCode]
	} else if dc.isRightDown() {
		dc.cachePoint = *dc.keyMap[RightKeyCode]
	}
	return &dc.cachePoint
}

func (dc *directionController) sendMouseEvent(controller Controller, repeat uint8) error {
	if dc.id == nil {
		if repeat != 0 {
			panic("repeat state error")
		}
		if dc.allUp() {
			panic("press state error")
		}

		dc.id = fingers.GetId()
		point := dc.getPoint(false)
		sme := singleMouseEvent{action: AMOTION_EVENT_ACTION_DOWN}
		sme.id = *dc.id
		sme.Point = *point
		return controller.PushEvent(&sme)
	} else {
		point := dc.getPoint(repeat != 0)
		sme := singleMouseEvent{}
		if dc.allUp() {
			sme.action = AMOTION_EVENT_ACTION_UP
		} else {
			sme.action = AMOTION_EVENT_ACTION_MOVE
		}
		sme.id = *dc.id
		sme.Point = *point
		b := controller.PushEvent(&sme)
		if dc.allUp() {
			fingers.Recycle(dc.id)
			dc.id = nil
		}
		return b
	}
}
