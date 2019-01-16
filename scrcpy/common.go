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

type Point struct {
	X uint16
	Y uint16
}

func (p Point) String() string {
	return fmt.Sprintf("Point: (%d, %d)", p.X, p.Y)
}

type position struct {
	screenSize size
	point      Point
}

//func (p position) Serialize(w io.Writer) error {
//	buf := make([]byte, 8)
//	ret := buf
//	binary.BigEndian.PutUint16(buf, p.Point.X)
//	buf = buf[2:]
//	binary.BigEndian.PutUint16(buf, p.Point.Y)
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
