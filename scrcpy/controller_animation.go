package scrcpy

import (
	"time"
)

type controllerAnimation struct {
	pointIntervals []*PointMacro
	state          int
	controller     Controller
	id             *int
	animator
}

var eventConstants = []androidMotionEventAction{AMOTION_EVENT_ACTION_DOWN, AMOTION_EVENT_ACTION_UP}

func newControllerAnimation(c Controller, pointIntervals []*PointMacro) *controllerAnimation {
	ca := controllerAnimation{
		pointIntervals: pointIntervals,
	}
	ca.InProgress = ca.inProgress
	ca.controller = c
	return &ca
}

func (ca *controllerAnimation) start() {
	ca.Start(nil)
}

func (ca *controllerAnimation) inProgress(data interface{}) time.Duration {
	m, n := ca.state/2, ca.state%2
	if m >= len(ca.pointIntervals) {
		panic("error state")
	}
	if n == 0 {
		ca.id = fingers.GetId()
	}
	sme := singleMouseEvent{action: eventConstants[n]}
	sme.id = *ca.id
	sme.Point = ca.pointIntervals[m].Point
	ca.controller.PushEvent(&sme)
	if n == 1 {
		fingers.Recycle(ca.id)
		ca.id = nil
	}
	ca.state++
	if n < 1 {
		return 30 * time.Millisecond
	} else if m == len(ca.pointIntervals)-1 {
		return 0
	} else {
		return ca.pointIntervals[m].Interval
	}
}
