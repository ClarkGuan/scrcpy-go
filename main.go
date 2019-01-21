package main

import (
	"log"

	"github.com/ClarkGuan/scrcpy-go/scrcpy"
	"github.com/veandco/go-sdl2/sdl"
)

func main() {
	option := scrcpy.Option{
		Debug:   true,
		BitRate: 8000000,
		MaxSize: 0,
		Port:    27183,
		KeyMap: map[int]*scrcpy.Point{
			scrcpy.FireKeyCode:   {509, 86},
			scrcpy.VisionKeyCode: {1600, 600},
			scrcpy.FrontKeyCode:  {350, 695},
			scrcpy.BackKeyCode:   {350, 921},
			sdl.K_SPACE:          {1994, 745},
			sdl.K_c:              {1964, 978},
			sdl.K_LSHIFT:         {1775, 1000},
		},
	}
	log.Println(scrcpy.Main(&option))
}
