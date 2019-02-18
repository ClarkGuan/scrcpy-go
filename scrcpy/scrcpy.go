package scrcpy

import (
	"log"
	"runtime"

	"github.com/ClarkGuan/go-sdl2/sdl"
)

type Option struct {
	Serial      string
	Crop        string
	Port        int
	MaxSize     int
	BitRate     int
	Debug       DebugLevel
	KeyMap      map[int]UserOperation
	CtrlKeyMap  map[int]UserOperation
	MouseKeyMap map[uint8]UserOperation
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
	if debugOpt.Debug() {
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

	controller := newController(svr.deviceConn, &screen)
	controller.Start()

	looper := NewSdlEventLooper()

	fh := &frameHandler{screen: &screen, frames: &frames}
	looper.Register(fh)

	ch := newControlHandler(controller,
		opt.KeyMap,
		opt.CtrlKeyMap,
		opt.MouseKeyMap)
	looper.Register(ch)
	screen.addRendererFunc(ch)

	if err = looper.Loop(); err != nil {
		log.Println(err)
	}

	sdl.Quit()
	return
}
