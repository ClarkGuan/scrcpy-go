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
	if tmp.data[C.int(i)] == nil {
		return nil
	}
	p := (*reflect.SliceHeader)(unsafe.Pointer(&ret))
	p.Data = uintptr(unsafe.Pointer(tmp.data[C.int(i)]))
	lineSize := af.lineSize(i)
	p.Len = lineSize * af.height()
	p.Cap = p.Len
	return
}

func (af avFrame) lineSize(i int) int {
	tmp := (*C.AVFrame)(unsafe.Pointer(af))
	return int(tmp.linesize[C.int(i)])
}

func (af avFrame) copy(b []byte) []byte {
	buf1, buf2 := af.data(0), af.data(1)
	if buf1 == nil || buf2 == nil {
		return b
	}
	if len(b) < len(buf1)+len(buf2) {
		b = make([]byte, len(buf1)+len(buf2))
	}
	n := copy(b, buf1)
	copy(b[n:], buf2) // FIXME 有概率会 crash，暂不知原因
	return b
}

func (af avFrame) isEmpty() bool {
	tmp := (*C.AVFrame)(unsafe.Pointer(af))
	return tmp.data[0] == nil
}

type frame struct {
	decodingFrame          avFrame
	hardwareFrame          avFrame
	renderingFrame         avFrame
	renderingFrameConsumed bool
	mutex                  sync.Mutex
}

func (f *frame) Init() error {
	if f.decodingFrame = avFrame(unsafe.Pointer(C.av_frame_alloc())); f.decodingFrame == 0 {
		return errAVAlloc
	}

	if f.hardwareFrame = avFrame(unsafe.Pointer(C.av_frame_alloc())); f.hardwareFrame == 0 {
		f.decodingFrame.free()
		return errAVAlloc
	}

	if f.renderingFrame = avFrame(unsafe.Pointer(C.av_frame_alloc())); f.renderingFrame == 0 {
		f.decodingFrame.free()
		f.hardwareFrame.free()
		return errAVAlloc
	}

	f.renderingFrameConsumed = true
	return nil
}

func (f *frame) Close() error {
	f.decodingFrame.free()
	f.hardwareFrame.free()
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

//export goGetHardwareFrame
func goGetHardwareFrame() *C.AVFrame {
	d := gDecoder
	return (*C.AVFrame)(unsafe.Pointer(d.hardwareFrame))
}

//export goAvHwframeTransferData
func goAvHwframeTransferData() C.int {
	d := gDecoder
	if !d.decodingFrame.isEmpty() {
		if d.decodingFrame.width() != d.hardwareFrame.width() ||
			(d.decodingFrame.height() != d.hardwareFrame.height()) {
			d.decodingFrame.free()
			d.decodingFrame = avFrame(unsafe.Pointer(C.av_frame_alloc()))
		}
	}
	return C.av_hwframe_transfer_data((*C.AVFrame)(unsafe.Pointer(d.decodingFrame)),
		goGetHardwareFrame(), 0)
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
	s.bufs = frame.copy(s.bufs)
	if s.bufs == nil {
		return nil
	}
	return s.texture.Update(nil, s.bufs, frame.lineSize(0))
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
