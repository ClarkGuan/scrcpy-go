package scrcpy

import "time"

type animator struct {
	timer      *time.Timer
	OnStart    func(data interface{})
	OnStop     func(data interface{})
	InProgress func(data interface{}) time.Duration
}

func (a *animator) Start(data interface{}) {
	go a.start(data)
}

func (a *animator) start(data interface{}) {
	if a.OnStart != nil {
		a.OnStart(data)
	}
	if a.InProgress != nil {
		for {
			a.resetTimer(a.InProgress(data))
			if a.timer == nil {
				break
			} else {
				<-a.timer.C
			}
		}
	}
	if a.OnStop != nil {
		a.OnStop(data)
	}
}

func (a *animator) resetTimer(d time.Duration) {
	if d > 0 {
		if a.timer == nil {
			a.timer = time.NewTimer(d)
		} else {
			a.timer.Reset(d)
		}
	} else {
		if a.timer != nil {
			a.timer.Stop()
			a.timer = nil
		}
	}
}
