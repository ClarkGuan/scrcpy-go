package scrcpy

import "time"

type mirrorMotion struct {
	clickPoint    Point
	endFlingPoint Point
	id            *int
	animator
	state int
}

func newMirrorMotion(clickPoint, endFlingPoint Point) *mirrorMotion {
	return &mirrorMotion{
		clickPoint:    clickPoint,
		endFlingPoint: endFlingPoint,
	}
}

func (mm *mirrorMotion) Start(c Controller) {
	mm.InProgress = mm.inProgress
	mm.animator.Start(c)
}

func (mm *mirrorMotion) inProgress(data interface{}) time.Duration {
	c := data.(Controller)
	switch mm.state {
	case 0:
		mm.state++
		mm.id = fingers.GetId()
		mm.sendMouseEvent(c, AMOTION_EVENT_ACTION_DOWN, *mm.id, mm.clickPoint)
		return 50 * time.Millisecond

	case 1:
		mm.state++
		mm.sendMouseEvent(c, AMOTION_EVENT_ACTION_MOVE, *mm.id, mm.clickPoint)
		return 50 * time.Millisecond

	case 2:
		mm.state++
		mm.sendMouseEvent(c, AMOTION_EVENT_ACTION_UP, *mm.id, mm.clickPoint)
		fingers.Recycle(mm.id)
		mm.id = nil
		return 150 * time.Millisecond

	case 3:
		mm.state++
		mm.id = fingers.GetId()
		mm.sendMouseEvent(c, AMOTION_EVENT_ACTION_DOWN, *mm.id, mm.endFlingPoint)
		return 50 * time.Millisecond

	case 4:
		mm.state++
		mm.sendMouseEvent(c, AMOTION_EVENT_ACTION_MOVE, *mm.id, mm.endFlingPoint)
		return 50 * time.Millisecond

	case 5:
		mm.state++
		mm.sendMouseEvent(c, AMOTION_EVENT_ACTION_UP, *mm.id, mm.endFlingPoint)
		fingers.Recycle(mm.id)
		mm.id = nil
		return 50 * time.Millisecond

	case 6:
		mm.state++
		mm.id = fingers.GetId()
		mm.sendMouseEvent(c, AMOTION_EVENT_ACTION_DOWN, *mm.id, mm.clickPoint)
		return 50 * time.Millisecond

	case 7:
		mm.state++
		mm.sendMouseEvent(c, AMOTION_EVENT_ACTION_MOVE, *mm.id, mm.clickPoint)
		return 50 * time.Millisecond

	case 8:
		mm.state++
		mm.sendMouseEvent(c, AMOTION_EVENT_ACTION_UP, *mm.id, mm.clickPoint)
		fingers.Recycle(mm.id)
		mm.id = nil
		return 0
	}
	panic("Can't reach here")
}

func (mm *mirrorMotion) sendMouseEvent(c Controller, action androidMotionEventAction, id int, p Point) error {
	sme := singleMouseEvent{action: action}
	sme.id = id
	sme.Point = p
	return c.PushEvent(&sme)
}
