package scrcpy

import (
	"fmt"
)

var debugOpt = false

type size struct {
	width  uint16
	height uint16
}

func (s size) Center() Point {
	return Point{s.width >> 1, s.height >> 1}
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
