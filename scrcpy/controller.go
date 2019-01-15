package scrcpy

import (
	"errors"
	"io"
	"log"
	"runtime"
	"sync"
	"sync/atomic"
)

var errFullQueue = errors.New("full event queue")
var errStopped = errors.New("queue already stopped")

type controlEventType uint8

const (
	CONTROL_EVENT_TYPE_KEYCODE controlEventType = iota
	CONTROL_EVENT_TYPE_TEXT
	CONTROL_EVENT_TYPE_MOUSE
	CONTROL_EVENT_TYPE_SCROLL
	CONTROL_EVENT_TYPE_COMMAND
)

type ControlEvent interface {
	EventType() controlEventType
	Serialize(w io.Writer, data ...interface{}) error
}

type Controller interface {
	Start()
	Stop() error
	PushEvent(interface{}) error
	Register(ControlEventHandler)
	Remove(ControlEventHandler)
	Writer() io.Writer
	Data() []interface{}
}

type ControlEventHandler interface {
	HandleControlEvent(Controller, interface{}) interface{}
}

type controllerImpl struct {
	writer  io.Writer
	data    []interface{}
	ch      chan interface{}
	stopped int32

	handlers     []ControlEventHandler
	handlerMutex sync.Mutex
}

func newController(w io.Writer, data ...interface{}) Controller {
	c := controllerImpl{writer: w, data: data, ch: make(chan interface{}, 512)}
	return &c
}

func (c *controllerImpl) Start() {
	go c.run()
}

func (c *controllerImpl) Stop() error {
	return c.PushEvent(nil)
}

func (c *controllerImpl) Register(handler ControlEventHandler) {
	c.handlerMutex.Lock()
	defer c.handlerMutex.Unlock()

	for _, h := range c.handlers {
		if h == handler {
			return
		}
	}
	c.handlers = append(c.handlers, handler)
}

func (c *controllerImpl) Remove(handler ControlEventHandler) {
	c.handlerMutex.Lock()
	defer c.handlerMutex.Unlock()

	for i := range c.handlers {
		if c.handlers[i] == handler {
			c.handlers = append(c.handlers[:i], c.handlers[i+1:]...)
			return
		}
	}
}

func (c *controllerImpl) run() {
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

		c.handlerMutex.Lock()
		ignoreDefault := false
		tmp := event
		for _, h := range c.handlers {
			if tmp = h.HandleControlEvent(c, tmp); tmp == nil {
				ignoreDefault = true
				break
			}
		}
		c.handlerMutex.Unlock()
		if !ignoreDefault {
			defaultControlHandler(c, tmp)
		}
	}
}

func (c *controllerImpl) Writer() io.Writer {
	return c.writer
}

func (c *controllerImpl) Data() []interface{} {
	return c.data
}

func (c *controllerImpl) PushEvent(ev interface{}) error {
	for {
		st := atomic.LoadInt32(&c.stopped)
		if st == 1 {
			return errStopped
		}
		if atomic.CompareAndSwapInt32(&c.stopped, 0, 2) {
			defer atomic.StoreInt32(&c.stopped, 0)
			select {
			case c.ch <- ev:
				return nil
			default:
				return errFullQueue
			}
		}
	}
}

type ControlHandlerFunc func(Controller, interface{}) bool

func (f ControlHandlerFunc) Handle(c Controller, event interface{}) bool {
	return f(c, event)
}

func defaultControlHandler(c Controller, event interface{}) interface{} {
	if ce, ok := event.(ControlEvent); ok {
		if err := ce.Serialize(c.Writer(), c.Data()...); err != nil {
			log.Println(err)
		}
	}

	return nil
}
