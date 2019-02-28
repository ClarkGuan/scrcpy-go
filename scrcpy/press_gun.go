package scrcpy

import (
	"sync/atomic"
	"time"
)

// 压枪处理
type gunPressOpration struct {
	animator
	stopFlag int32
	delta    int32
	interval time.Duration
}

func (gpo *gunPressOpration) Start(c *visionController, interval time.Duration, delta int) {
	gpo.animator.InProgress = gpo.inProgress
	gpo.SetValues(interval, delta)
	gpo.animator.Start(c)
}

func (gpo *gunPressOpration) SetValues(interval time.Duration, delta int) {
	if interval < 30*time.Millisecond {
		gpo.interval = 30 * time.Millisecond
	} else {
		gpo.interval = interval
	}
	gpo.delta = int32(delta)
}

func (gpo *gunPressOpration) inProgress(data interface{}) time.Duration {
	controller := data.(*visionController)
	if atomic.LoadInt32(&gpo.stopFlag) != 1 {
		controller.visionControl2(0, gpo.delta)
		return gpo.interval
	}
	return 0
}

func (gpo *gunPressOpration) Stop() {
	atomic.StoreInt32(&gpo.stopFlag, 1)
}
