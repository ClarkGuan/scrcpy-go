package scrcpy

import "github.com/ClarkGuan/go-sdl2/sdl"

const (
	BUTTON_LEFT               = "BUTTON_LEFT"
	BUTTON_MIDDLE             = "BUTTON_MIDDLE"
	BUTTON_RIGHT              = "BUTTON_RIGHT"
	BUTTON_X1                 = "BUTTON_X1"
	BUTTON_X2                 = "BUTTON_X2"
	SCRCPY_FIRE               = "SCRCPY_FIRE"
	SCRCPY_VISION_TOPLEFT     = "SCRCPY_VISION_TOPLEFT"
	SCRCPY_VISION_BOTTOMRIGHT = "SCRCPY_VISION_BOTTOMRIGHT"
	SCRCPY_FRONT              = "SCRCPY_FRONT"
	SCRCPY_BACK               = "SCRCPY_BACK"
)

var MouseButtonMap = map[string]uint8{
	BUTTON_LEFT:   sdl.BUTTON_LEFT,
	BUTTON_MIDDLE: sdl.BUTTON_MIDDLE,
	BUTTON_RIGHT:  sdl.BUTTON_RIGHT,
	BUTTON_X1:     sdl.BUTTON_X1,
	BUTTON_X2:     sdl.BUTTON_X2,
}

var KeyCodeConstMap = map[string]int{
	SCRCPY_FIRE:               FireKeyCode,
	SCRCPY_VISION_TOPLEFT:     VisionBoundTopLeft,
	SCRCPY_VISION_BOTTOMRIGHT: VisionBoundBottomRight,
	SCRCPY_FRONT:              FrontKeyCode,
	SCRCPY_BACK:               BackKeyCode,
}
