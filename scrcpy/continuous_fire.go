package scrcpy

import (
	"sync/atomic"
	"time"
)

type continuousFire struct {
	animator
	stopFlag int32
	Point
	state    int
	id       *int
	interval time.Duration
}

func (cf *continuousFire) Start(c Controller, interval time.Duration) {
	cf.animator.InProgress = cf.inProgress
	cf.SetInterval(interval)
	cf.animator.Start(c)
}

func (cf *continuousFire) SetInterval(interval time.Duration) {
	if interval < 10*time.Millisecond {
		cf.interval = 10 * time.Millisecond
	} else {
		cf.interval = interval
	}
}

func (cf *continuousFire) inProgress(data interface{}) time.Duration {
	c := data.(Controller)
	if atomic.LoadInt32(&cf.stopFlag) != 1 {
		cf.state = cf.state % 2
		switch cf.state {
		case 0:
			cf.id = fingers.GetId()
			cf.sendMouseEvent(c, AMOTION_EVENT_ACTION_DOWN, *cf.id)
			cf.state++
			return cf.interval

		case 1:
			cf.sendMouseEvent(c, AMOTION_EVENT_ACTION_UP, *cf.id)
			fingers.Recycle(cf.id)
			cf.id = nil
			cf.state++
			return cf.interval
		}
		panic("can't reach here")
	} else {
		if cf.id != nil {
			cf.sendMouseEvent(c, AMOTION_EVENT_ACTION_UP, *cf.id)
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
