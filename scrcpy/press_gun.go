package scrcpy

import (
	"sync/atomic"
	"time"
)

// 压枪处理
type gunPressOperation struct {
	animator
	stopFlag int32
	gunPressConfig
}

type gunPressConfig struct {
	delta    int32
	interval time.Duration
}

func (gpo *gunPressOperation) Start(c *visionController, config gunPressConfig) {
	gpo.animator.InProgress = gpo.inProgress
	gpo.SetValues(config)
	gpo.animator.Start(c)
}

func (gpo *gunPressOperation) SetValues(config gunPressConfig) {
	if config.interval < 10*time.Millisecond {
		gpo.interval = 10 * time.Millisecond
	} else {
		gpo.interval = config.interval
	}
	gpo.delta = int32(config.delta)
}

func (gpo *gunPressOperation) inProgress(data interface{}) time.Duration {
	controller := data.(*visionController)
	if atomic.LoadInt32(&gpo.stopFlag) != 1 {
		controller.visionControl2(0, gpo.delta)
		return gpo.interval
	}
	return 0
}

func (gpo *gunPressOperation) Stop() {
	atomic.StoreInt32(&gpo.stopFlag, 1)
}
