package scrcpy

import (
	"log"
	"runtime"

	"github.com/veandco/go-sdl2/sdl"
)

type Option struct {
	Serial  string
	Crop    string
	Port    int
	MaxSize int
	BitRate int
	Debug   bool
}

func Main(opt *Option) (err error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	debugOpt = opt.Debug

	svr := server{}
	svrOpt := serverOption{serial: opt.Serial, localPort: opt.Port,
		maxSize: opt.MaxSize, bitRate: opt.BitRate, crop: opt.Crop}

	if err = svr.Start(&svrOpt); err != nil {
		return
	}
	defer func() {
		svr.Stop()
		svr.Close()
	}()

	if err = sdlInitAndConfigure(); err != nil {
		return
	}

	if err = svr.ConnectTo(); err != nil {
		return
	}

	var deviceName string
	var screenSize size
	if deviceName, screenSize, err = svr.ReadDeviceInfo(); err != nil {
		return
	}
	if debugOpt {
		log.Printf("device name: %s, screen %v\n", deviceName, screenSize)
	}

	frames := frame{}
	if err = frames.Init(); err != nil {
		return
	}
	defer frames.Close()

	decoder := getDecoder(&frames, svr.deviceConn)
	decoder.Start()

	screen := screen{}
	if err = screen.InitRendering(deviceName, screenSize); err != nil {
		return
	}

	controller := newController(&screen, svr.deviceConn)
	controller.Start()

	if err = eventLoop(&screen, &frames, controller); err != nil {
		log.Println(err)
	}

	return
}

func eventLoop(screen *screen, frames *frame, c *controller) error {
	var ev sdl.Event
	for ev = sdl.WaitEvent(); ev != nil; ev = sdl.WaitEvent() {
		switch ev.GetType() {
		case eventNewFrame:
			if !screen.hasFrame {
				screen.hasFrame = true
				screen.showWindow()
			}
			if err := screen.updateFrame(frames); err != nil {
				return err
			}

		case eventDecoderStopped:
			log.Println("Video decoder stopped")
			return nil

		case sdl.QUIT:
			log.Println("User requested to quit")
			return nil

		case sdl.MOUSEMOTION:
			mme := ev.(*sdl.MouseMotionEvent)
			if mme.State == 0 {
				continue
			}
			sme := singleMouseEvent{}
			sme.action = AMOTION_EVENT_ACTION_MOVE
			sme.id = 0
			sme.point.x = uint16(mme.X)
			sme.point.y = uint16(mme.Y)
			c.PushEvent(&sme)

		case sdl.MOUSEBUTTONDOWN:
			fallthrough
		case sdl.MOUSEBUTTONUP:
			mbe := ev.(*sdl.MouseButtonEvent)
			sme := singleMouseEvent{}
			if ev.GetType() == sdl.MOUSEBUTTONDOWN {
				sme.action = AMOTION_EVENT_ACTION_DOWN
			} else {
				sme.action = AMOTION_EVENT_ACTION_UP
			}
			sme.point.x = uint16(mbe.X)
			sme.point.y = uint16(mbe.Y)
			c.PushEvent(&sme)

		case sdl.KEYDOWN:
			fallthrough
		case sdl.KEYUP:
			//kbe := ev.(*sdl.KeyboardEvent)
		}
	}
	return nil
}
