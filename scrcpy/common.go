package scrcpy

import (
	"fmt"
)

type DebugLevel int

const (
	MinLevel DebugLevel = iota
	Error
	Warn
	Info
	Debug
	MaxLevel
)

func (dl DebugLevel) Debug() bool {
	return dl >= Debug
}

func (dl DebugLevel) Info() bool {
	return dl >= Info
}

func (dl DebugLevel) Warn() bool {
	return dl >= Warn
}

func (dl DebugLevel) Error() bool {
	return dl >= Error
}

var debugOpt = MinLevel

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
