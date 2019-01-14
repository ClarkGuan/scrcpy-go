package scrcpy

import (
	"errors"
	"io"
	"log"
	"net"
	"runtime"
	"sync"
	"sync/atomic"
)

var errFullQueue = errors.New("full event queue")
var errStopped = errors.New("queue already stopped")

type controller struct {
	videoSock net.Conn
	screen    *screen
	ch        chan interface{}
	stopped   int32

	mouseEvents mouseEventSet
	mainTouchId int
}

func newController(screen *screen, sock net.Conn) *controller {
	c := controller{screen: screen, videoSock: sock, ch: make(chan interface{}, 512)}
	return &c
}

func (c *controller) Start() {
	go c.run()
}

func (c *controller) run() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	for {
		event := <-c.ch
		if event == nil {
			for {
				st := atomic.LoadInt32(&c.stopped)
				if st != 0 {
					continue
				}
				if atomic.CompareAndSwapInt32(&c.stopped, 0, 1) {
					close(c.ch)
				}
			}
			break
		}

		switch ce := event.(type) {
		case controlEvent:
			if err := ce.Serialize(c.videoSock, c.screen); err != nil {
				log.Println(err)
				return
			}

		case *singleMouseEvent:
			c.mouseEvents.accept(ce)
			if err := c.mouseEvents.Serialize(c.videoSock, c.screen); err != nil {
				log.Println(err)
				return
			}
		}

	}
}

func (c *controller) PushEvent(ev interface{}) error {
	for {
		st := atomic.LoadInt32(&c.stopped)
		if st == 1 {
			return errStopped
		}
		if atomic.CompareAndSwapInt32(&c.stopped, 0, 2) {
			defer atomic.StoreInt32(&c.stopped, 0)
			select {
			case c.ch <- ev:
				// 此时就可以将此 id 看做回收了
				if sme, ok := ev.(*singleMouseEvent); ok && sme.action == AMOTION_EVENT_ACTION_UP {
					c.mouseEvents.table[sme.id] = false
				}
				return nil
			default:
				return errFullQueue
			}
		}
	}
}

func (c *controller) ReleaseAllTouch() error {
	c.mouseEvents.Lock()
	defer c.mouseEvents.Unlock()

	for id, p := range c.mouseEvents.points {
		event := singleMouseEvent{action: AMOTION_EVENT_ACTION_UP}
		event.point = p
		event.id = id
		if err := c.PushEvent(&event); err != nil {
			return err
		}
	}
	return nil
}

func (c *controller) Stop() error {
	return c.PushEvent(nil)
}

type controlEventType uint8

const (
	CONTROL_EVENT_TYPE_KEYCODE controlEventType = iota
	CONTROL_EVENT_TYPE_TEXT
	CONTROL_EVENT_TYPE_MOUSE
	CONTROL_EVENT_TYPE_SCROLL
	CONTROL_EVENT_TYPE_COMMAND
)

type controlEvent interface {
	EventType() controlEventType
	Serialize(w io.Writer, s *screen) error
}

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
	point
	id int
}

// 多点触摸，每一个点一旦 down，就会生成一个 id，且该 id 在 up 之前不变
type mouseEventSet struct {
	sync.Mutex
	points map[int]point
	buf    []byte
	action androidMotionEventAction
	id     int

	// SDL 事件循环线程访问
	table [128]bool
}

func (set *mouseEventSet) acquireId() int {
	for i := range set.table {
		if !set.table[i] {
			return i
		}
	}
	panic("out of touch count")
}

func (set *mouseEventSet) accept(se *singleMouseEvent) {
	set.Lock()
	if set.points == nil {
		set.points = make(map[int]point)
	}
	set.points[se.id] = se.point
	set.Unlock()

	if se.action == AMOTION_EVENT_ACTION_DOWN && se.id != 0 {
		se.action = AMOTION_EVENT_ACTION_POINTER_DOWN | androidMotionEventAction(se.id)<<8
	} else if se.action == AMOTION_EVENT_ACTION_UP && len(set.points) > 1 {
		se.action = AMOTION_EVENT_ACTION_POINTER_UP | androidMotionEventAction(1<<8)
	}
	set.action = se.action
	set.id = se.id
}

func (set *mouseEventSet) Serialize(w io.Writer, s *screen) error {
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

	set.Lock()
	defer set.Unlock()
	// 写入数组长度 1 个字节
	set.buf = append(set.buf, byte(len(set.points)))

	// 写入数组内容
	for id, p := range set.points {
		set.buf = append(set.buf, byte(p.x>>8))
		set.buf = append(set.buf, byte(p.x))
		set.buf = append(set.buf, byte(p.y>>8))
		set.buf = append(set.buf, byte(p.y))
		set.buf = append(set.buf, byte(id))
	}

	// 写入 frame size
	set.buf = append(set.buf, byte(s.frameSize.width>>8))
	set.buf = append(set.buf, byte(s.frameSize.width))
	set.buf = append(set.buf, byte(s.frameSize.height>>8))
	set.buf = append(set.buf, byte(s.frameSize.height))

	_, err := w.Write(set.buf)

	if set.action == AMOTION_EVENT_ACTION_UP || set.action == AMOTION_EVENT_ACTION_POINTER_UP {
		delete(set.points, set.id)
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
