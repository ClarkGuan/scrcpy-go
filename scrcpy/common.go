package scrcpy

import (
	"fmt"
)

var debugOpt = false

type size struct {
	width  uint16
	height uint16
}

func (s size) String() string {
	return fmt.Sprintf("size: (%d, %d)", s.width, s.height)
}

type point struct {
	x uint16
	y uint16
}

func (p point) String() string {
	return fmt.Sprintf("point: (%d, %d)", p.x, p.y)
}

type position struct {
	screenSize size
	point      point
}

//func (p position) Serialize(w io.Writer) error {
//	buf := make([]byte, 8)
//	ret := buf
//	binary.BigEndian.PutUint16(buf, p.point.x)
//	buf = buf[2:]
//	binary.BigEndian.PutUint16(buf, p.point.y)
//	buf = buf[2:]
//	binary.BigEndian.PutUint16(buf, p.screenSize.width)
//	buf = buf[2:]
//	binary.BigEndian.PutUint16(buf, p.screenSize.height)
//	_, err := w.Write(ret)
//	return err
//}

func (p position) String() string {
	return fmt.Sprintf("position: {%v, %v}", p.screenSize, p.point)
}
