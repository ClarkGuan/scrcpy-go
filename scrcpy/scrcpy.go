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

	sdl.Quit()
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
			processMouseMotionEvent(ev.(*sdl.MouseMotionEvent), c)

		case sdl.MOUSEBUTTONDOWN:
			fallthrough
		case sdl.MOUSEBUTTONUP:
			processMouseButtonEvent(ev.(*sdl.MouseButtonEvent), c)

		case sdl.KEYDOWN:
			fallthrough
		case sdl.KEYUP:
			processKeyboardEvent(ev.(*sdl.KeyboardEvent), c)
		}
	}
	return nil
}

func processMouseMotionEvent(mme *sdl.MouseMotionEvent, c *controller) {
	if !sdl.GetRelativeMouseMode() && mme.State == 0 {
		return
	}

	sme := singleMouseEvent{}
	sme.action = AMOTION_EVENT_ACTION_MOVE
	sme.id = c.mainTouchId
	sme.point.x = uint16(mme.X)
	sme.point.y = uint16(mme.Y)
	c.PushEvent(&sme)
}

func processMouseButtonEvent(mbe *sdl.MouseButtonEvent, c *controller) {
	if sdl.GetRelativeMouseMode() {
		// TODO 射击
	} else {
		sme := singleMouseEvent{}
		if mbe.Type == sdl.MOUSEBUTTONDOWN {
			sme.action = AMOTION_EVENT_ACTION_DOWN
			c.mainTouchId = c.mouseEvents.acquireId()
		} else {
			sme.action = AMOTION_EVENT_ACTION_UP
		}
		sme.id = c.mainTouchId
		sme.point.x = uint16(mbe.X)
		sme.point.y = uint16(mbe.Y)
		c.PushEvent(&sme)
	}
}

func processKeyboardEvent(kbe *sdl.KeyboardEvent, c *controller) {
	ctrl := kbe.Keysym.Mod&(sdl.KMOD_RCTRL|sdl.KMOD_LCTRL) != 0
	alt := kbe.Keysym.Mod&(sdl.KMOD_RALT|sdl.KMOD_LALT) != 0
	//meta := kbe.Keysym.Mod & (sdl.KMOD_RGUI | sdl.KMOD_LGUI) != 0  // command on mac

	if alt {
		return
	}

	keycode := kbe.Keysym.Sym
	if ctrl && keycode == sdl.K_x && kbe.Type == sdl.KEYUP {
		isMouseRelativeMode := sdl.GetRelativeMouseMode()
		sdl.SetRelativeMouseMode(!isMouseRelativeMode)
		c.ReleaseAllTouch()

		if !isMouseRelativeMode {
			if debugOpt {
				log.Println("切换鼠标模式为射击模式")
			}

			c.mainTouchId = c.mouseEvents.acquireId()
			event := singleMouseEvent{action: AMOTION_EVENT_ACTION_DOWN}
			event.x = c.screen.frameSize.width >> 1
			event.y = c.screen.frameSize.height >> 1
			event.id = c.mainTouchId
			c.PushEvent(&event)
		} else {
			if debugOpt {
				log.Println("切换鼠标模式为一般模式")
			}
		}
	}
}
