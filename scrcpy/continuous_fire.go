package scrcpy

import (
	"sync/atomic"
	"time"
)

type continuousFire struct {
	animator
	stopFlag int32
	Point
	state int
	id    *int
	up    bool
}

func (cf *continuousFire) Start(c Controller) {
	cf.animator.InProgress = cf.inProgress
	cf.animator.Start(c)
}

func (cf *continuousFire) inProgress(data interface{}) time.Duration {
	c := data.(Controller)
	if atomic.LoadInt32(&cf.stopFlag) != 1 {
		cf.state = cf.state % 3
		switch cf.state {
		case 0:
			if cf.id == nil {
				cf.id = fingers.GetId()
			}
			cf.up = false
			cf.sendMouseEvent(c, AMOTION_EVENT_ACTION_DOWN, *cf.id)

		case 1:
			cf.sendMouseEvent(c, AMOTION_EVENT_ACTION_MOVE, *cf.id)
			cf.up = false

		case 2:
			cf.sendMouseEvent(c, AMOTION_EVENT_ACTION_UP, *cf.id)
			cf.up = true

		default:
			panic("can't reach here")
		}
		cf.state++
		return 30 * time.Millisecond
	} else {
		if cf.id != nil {
			if !cf.up {
				cf.sendMouseEvent(c, AMOTION_EVENT_ACTION_UP, *cf.id)
				cf.up = true
			}
			fingers.Recycle(cf.id)
			cf.id = nil
		}
		return 0
	}
}

func (cf *continuousFire) Stop() {
	atomic.StoreInt32(&cf.stopFlag, 1)
}

func (cf *continuousFire) sendMouseEvent(c Controller, action androidMotionEventAction, id int) error {
	sme := singleMouseEvent{action: action}
	sme.id = id
	sme.Point = cf.Point
	return c.PushEvent(&sme)
}
