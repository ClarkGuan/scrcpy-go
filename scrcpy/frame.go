package scrcpy

//
// #cgo LDFLAGS: -L/usr/local/lib -lavformat -lavcodec -lavutil
// #cgo CFLAGS: -I/usr/local/include
//
// #include <libavutil/avutil.h>
// #include <libavformat/avformat.h>
//
// int run_decoder();
//
import "C"
import (
	"errors"
	"io"
	"log"
	"net"
	"reflect"
	"runtime"
	"sync"
	"unsafe"

	"github.com/ClarkGuan/go-sdl2/sdl"
)

const eventNewFrame = sdl.USEREVENT + 1
const eventDecoderStopped = sdl.USEREVENT + 2

var errAVAlloc = errors.New("av_frame_alloc() fail")

type avFrame uintptr

func (af avFrame) width() int {
	return int((*C.AVFrame)(unsafe.Pointer(af)).width)
}

func (af avFrame) height() int {
	return int((*C.AVFrame)(unsafe.Pointer(af)).height)
}

func (af avFrame) free() {
	tmp := (*C.AVFrame)(unsafe.Pointer(af))
	C.av_frame_free(&tmp)
}

func (af avFrame) data(i int) (ret []byte) {
	tmp := (*C.AVFrame)(unsafe.Pointer(af))
	p := (*reflect.SliceHeader)(unsafe.Pointer(&ret))
	p.Data = uintptr(unsafe.Pointer(tmp.data[C.int(i)]))
	p.Len = 1
	p.Cap = 1
	return
}

func (af avFrame) lineSize(i int) int {
	tmp := (*C.AVFrame)(unsafe.Pointer(af))
	return int(tmp.linesize[C.int(i)])
}

type frame struct {
	decodingFrame          avFrame
	renderingFrame         avFrame
	renderingFrameConsumed bool
	mutex                  sync.Mutex
}

func (f *frame) Init() error {
	if f.decodingFrame = avFrame(unsafe.Pointer(C.av_frame_alloc())); f.decodingFrame == 0 {
		return errAVAlloc
	}

	if f.renderingFrame = avFrame(unsafe.Pointer(C.av_frame_alloc())); f.renderingFrame == 0 {
		f.decodingFrame.free()
		return errAVAlloc
	}

	f.renderingFrameConsumed = true
	return nil
}

func (f *frame) Close() error {
	f.decodingFrame.free()
	f.renderingFrame.free()
	return nil
}

func (f *frame) swap() {
	tmp := f.renderingFrame
	f.renderingFrame = f.decodingFrame
	f.decodingFrame = tmp
}

func (f *frame) OfferDecodeFrame() bool {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	f.swap()
	prev := f.renderingFrameConsumed
	f.renderingFrameConsumed = false
	return prev
}

func (f *frame) ConsumeRenderedFrame() avFrame {
	if f.renderingFrameConsumed {
		panic(errors.New("renderingFrameConsumed state error"))
	}

	f.renderingFrameConsumed = true
	return f.renderingFrame
}

type decoder struct {
	*frame
	videoSock net.Conn
}

var gDecoder *decoder

func getDecoder(f *frame, sock net.Conn) *decoder {
	if gDecoder == nil {
		gDecoder = &decoder{frame: f, videoSock: sock}
	}
	return gDecoder
}

func (d *decoder) Start() {
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		C.run_decoder()
	}()
}

//export goPushFrame
func goPushFrame() {
	d := gDecoder
	previousFrameConsumed := d.OfferDecodeFrame()
	if !previousFrameConsumed {
		return
	}
	// 使用自定义 SDL 事件关联
	sdl.PushEvent(&sdl.UserEvent{Type: eventNewFrame})
}

//export goNotifyStopped
func goNotifyStopped() {
	sdl.PushEvent(&sdl.UserEvent{Type: eventDecoderStopped})
}

//export goGetDecodingFrame
func goGetDecodingFrame() *C.AVFrame {
	d := gDecoder
	return (*C.AVFrame)(unsafe.Pointer(d.decodingFrame))
}

//export goReadPacket
func goReadPacket(_, buf unsafe.Pointer, bufSize C.int) C.int {
	d := gDecoder
	var buffer []byte
	pb := (*reflect.SliceHeader)(unsafe.Pointer(&buffer))
	pb.Data = uintptr(buf)
	pb.Cap = int(bufSize)
	pb.Len = int(bufSize)

	if n, err := d.videoSock.Read(buffer); err == io.EOF {
		return 0
	} else if err != nil {
		return -1
	} else {
		return C.int(n)
	}
}

func (s *screen) updateFrame(frames *frame) error {
	frames.mutex.Lock()

	frame := frames.ConsumeRenderedFrame()
	if err := s.prepareForFrame(size{width: uint16(frame.width()),
		height: uint16(frame.height())}); err != nil {
		frames.mutex.Unlock()
		return err
	}

	if err := s.updateTexture(frame); err != nil {
		frames.mutex.Unlock()
		return err
	}
	frames.mutex.Unlock()
	s.render()
	return nil
}

func (s *screen) updateTexture(frame avFrame) error {
	return s.texture.UpdateYUV(nil,
		frame.data(0), frame.lineSize(0),
		frame.data(1), frame.lineSize(1),
		frame.data(2), frame.lineSize(2))
}

type frameHandler struct {
	screen *screen
	frames *frame
}

func (fh *frameHandler) HandleSdlEvent(event sdl.Event) (bool, error) {
	switch event.GetType() {
	case eventNewFrame:
		if !fh.screen.hasFrame {
			fh.screen.hasFrame = true
			fh.screen.showWindow()
		}
		if err := fh.screen.updateFrame(fh.frames); err != nil {
			return true, err
		}

	case eventDecoderStopped:
		log.Println("Video decoder stopped")
		return true, nil
	}

	return false, nil
}
