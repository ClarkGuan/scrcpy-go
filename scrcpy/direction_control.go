package scrcpy

type Direction int

const (
	frontDirection Direction = 1 << iota
	backDirection
	leftDirection
	rightDirection
)

const deltaDirectionMovement = 150

type directionController struct {
	direction   Direction
	cachePoint  Point
	middlePoint *Point
	radius      uint16
	keyMap      map[int]*Point
	id          *int
	lastPoint   Point
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
		dc.middlePoint.X = dc.keyMap[FrontKeyCode].X
		dc.middlePoint.Y = (dc.keyMap[FrontKeyCode].Y + dc.keyMap[BackKeyCode].Y) >> 1
		dc.radius = dc.middlePoint.Y - dc.keyMap[FrontKeyCode].Y
	}
}

func (dc *directionController) getPoint(repeat bool) *Point {
	dc.prepare()
	dc.cachePoint = *dc.middlePoint

	//if debugOpt {
	//	log.Println("中间点：", *dc.middlePoint, "半径：", dc.radius)
	//}

	if dc.isFrontDown() {
		//if debugOpt {
		//	log.Println("向前")
		//}
		dc.cachePoint.Y -= dc.radius
	}

	if dc.isLeftDown() {
		//if debugOpt {
		//	log.Println("向左")
		//}
		dc.cachePoint.X -= dc.radius
	}

	if dc.isRightDown() {
		//if debugOpt {
		//	log.Println("向右")
		//}
		dc.cachePoint.X += dc.radius
	}

	if dc.isBackDown() {
		//if debugOpt {
		//	log.Println("向后")
		//}
		dc.cachePoint.Y += dc.radius
	}

	if dc.cachePoint.Y < dc.middlePoint.Y {
		if repeat {
			//if debugOpt {
			//	log.Println("向前跑")
			//}
			dc.cachePoint.Y -= deltaDirectionMovement
		}

		if dc.cachePoint.X < dc.middlePoint.X {
			//if debugOpt {
			//	log.Println("向左跑")
			//}
			dc.cachePoint.X -= deltaDirectionMovement
		} else if dc.cachePoint.X > dc.middlePoint.X {
			//if debugOpt {
			//	log.Println("向右跑")
			//}
			dc.cachePoint.X += deltaDirectionMovement
		}
	}

	return &dc.cachePoint
}

func (dc *directionController) sendMouseEvent(controller Controller) error {
	if dc.id == nil {
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
		point := &dc.cachePoint
		sme := singleMouseEvent{}
		if dc.allUp() {
			sme.action = AMOTION_EVENT_ACTION_UP
			dc.lastPoint.Y = 0
			dc.lastPoint.X = 0
		} else {
			sme.action = AMOTION_EVENT_ACTION_MOVE
			point = dc.getPoint(true)

			if dc.lastPoint == *point {
				// 优化，少发一些事件
				return nil
			} else {
				dc.lastPoint = *point
			}
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
