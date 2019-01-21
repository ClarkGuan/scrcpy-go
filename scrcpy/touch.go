package scrcpy

import (
	"io"
)

// touch pointer 规则：
// 1、id 在 touch 过程中不变
// 2、index 在 touch 过程中可能会改变，因为可能数组中间某个 pointer 先松手
// 3、id 小的一定排前面，id 大的排后面
// 4、中途松开的 id 再次按下时可以复用之前的 id

type androidMotionEventAction uint16

const (
	AMOTION_EVENT_ACTION_MASK               androidMotionEventAction = 0xff
	AMOTION_EVENT_ACTION_POINTER_INDEX_MASK androidMotionEventAction = 0xff00
	AMOTION_EVENT_ACTION_DOWN               androidMotionEventAction = 0
	AMOTION_EVENT_ACTION_UP                 androidMotionEventAction = 1
	AMOTION_EVENT_ACTION_MOVE               androidMotionEventAction = 2
	AMOTION_EVENT_ACTION_CANCEL             androidMotionEventAction = 3
	AMOTION_EVENT_ACTION_OUTSIDE            androidMotionEventAction = 4
	AMOTION_EVENT_ACTION_POINTER_DOWN       androidMotionEventAction = 5
	AMOTION_EVENT_ACTION_POINTER_UP         androidMotionEventAction = 6
	AMOTION_EVENT_ACTION_HOVER_MOVE         androidMotionEventAction = 7
	AMOTION_EVENT_ACTION_SCROLL             androidMotionEventAction = 8
	AMOTION_EVENT_ACTION_HOVER_ENTER        androidMotionEventAction = 9
	AMOTION_EVENT_ACTION_HOVER_EXIT         androidMotionEventAction = 10
	AMOTION_EVENT_ACTION_BUTTON_PRESS       androidMotionEventAction = 11
	AMOTION_EVENT_ACTION_BUTTON_RELEASE     androidMotionEventAction = 12
)

type touchPoint struct {
	Point
	id int
}

// 多点触摸，每一个点一旦 down，就会生成一个 id，且该 id 在 up 之前不变
type mouseEventSet struct {
	points []touchPoint
	buf    []byte

	action androidMotionEventAction
	id     int
}

func swapPoint(a, b *touchPoint) {
	tmp := a.X
	a.X = b.X
	b.X = tmp
	tmp = a.Y
	a.Y = b.Y
	b.Y = tmp
	id := a.id
	a.id = b.id
	b.id = id
}

func (set *mouseEventSet) accept(se *singleMouseEvent) {
	index := -1
	if se.action == AMOTION_EVENT_ACTION_DOWN {
		set.points = append(set.points, touchPoint{Point: se.Point, id: se.id})
		var i int
		for i = len(set.points) - 1; i >= 1; i-- {
			if set.points[i].id < set.points[i-1].id {
				swapPoint(&set.points[i], &set.points[i-1])
			} else {
				break
			}
		}
		index = i
	} else {
		for i := range set.points {
			if set.points[i].id == se.id {
				set.points[i].Point = se.Point
				index = i
			}
		}
	}
	if index == -1 {
		panic("pointer not found")
	}

	if se.action == AMOTION_EVENT_ACTION_DOWN && len(set.points) > 1 {
		se.action = AMOTION_EVENT_ACTION_POINTER_DOWN | androidMotionEventAction(index)<<8
	} else if se.action == AMOTION_EVENT_ACTION_UP && len(set.points) > 1 {
		se.action = AMOTION_EVENT_ACTION_POINTER_UP | androidMotionEventAction(index)<<8
	}
	set.action = se.action
	set.id = se.id
}

func (set *mouseEventSet) Serialize(w io.Writer, data ...interface{}) error {
	if set.buf == nil {
		set.buf = make([]byte, 0, 128)
	} else {
		set.buf = set.buf[:0]
	}

	// 写入 type
	set.buf = append(set.buf, byte(set.EventType()))

	// 写入 action
	set.buf = append(set.buf, byte(set.action>>8))
	set.buf = append(set.buf, byte(set.action))

	// 写入数组长度 1 个字节
	set.buf = append(set.buf, byte(len(set.points)))

	// 写入数组内容
	for _, p := range set.points {
		set.buf = append(set.buf, byte(p.X>>8))
		set.buf = append(set.buf, byte(p.X))
		set.buf = append(set.buf, byte(p.Y>>8))
		set.buf = append(set.buf, byte(p.Y))
		set.buf = append(set.buf, byte(p.id))
	}

	// 写入 frame size
	s := data[0].(*screen)
	set.buf = append(set.buf, byte(s.frameSize.width>>8))
	set.buf = append(set.buf, byte(s.frameSize.width))
	set.buf = append(set.buf, byte(s.frameSize.height>>8))
	set.buf = append(set.buf, byte(s.frameSize.height))

	_, err := w.Write(set.buf)

	if set.action == AMOTION_EVENT_ACTION_UP {
		set.points = set.points[:0]
	} else if (set.action & 0x00ff) == AMOTION_EVENT_ACTION_POINTER_UP {
		index := (set.action & 0xff00) >> 8
		set.points = append(set.points[:index], set.points[index+1:]...)
	}

	return err
}

func (set *mouseEventSet) EventType() controlEventType {
	return CONTROL_EVENT_TYPE_MOUSE
}

type singleMouseEvent struct {
	touchPoint
	action androidMotionEventAction
}

type fingerState [8]bool

var fingers fingerState

func (f *fingerState) GetId() *int {
	for i := range f[:] {
		if !f[i] {
			f[i] = true
			return &i
		}
	}
	panic("finger number over 8")
}

func (f *fingerState) Recycle(i *int) {
	f[*i] = false
}
