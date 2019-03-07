package scrcpy

import (
	"fmt"
	"sync/atomic"
	"time"
)

// 压枪处理
type gunPressOperation struct {
	animator
	stopFlag int32
	GunPressConfig
}

type GunPressConfig struct {
	Delta    int32
	Interval time.Duration
}

func (c *GunPressConfig) String() string {
	return fmt.Sprintf("(%d, %s)", c.Delta, c.Interval)
}

func (gpo *gunPressOperation) Start(c *visionController, config GunPressConfig) {
	gpo.animator.InProgress = gpo.inProgress
	gpo.SetValues(config)
	gpo.animator.Start(c)
}

func (gpo *gunPressOperation) SetValues(config GunPressConfig) {
	if config.Interval < 10*time.Millisecond {
		gpo.Interval = 10 * time.Millisecond
	} else {
		gpo.Interval = config.Interval
	}
	gpo.Delta = int32(config.Delta)
}

func (gpo *gunPressOperation) inProgress(data interface{}) time.Duration {
	controller := data.(*visionController)
	if atomic.LoadInt32(&gpo.stopFlag) != 1 {
		controller.visionControl2(0, gpo.Delta)
		return gpo.Interval
	}
	return 0
}

func (gpo *gunPressOperation) Stop() {
	atomic.StoreInt32(&gpo.stopFlag, 1)
}
