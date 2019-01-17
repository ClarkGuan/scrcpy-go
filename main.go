package main

import (
	"log"

	"github.com/ClarkGuan/scrcpy-go/scrcpy"
	"github.com/veandco/go-sdl2/sdl"
)

func main() {
	option := scrcpy.Option{
		Debug:   false,
		BitRate: 8000000,
		MaxSize: 0,
		Port:    27183,
		KeyMap: map[int]*scrcpy.Point{
			scrcpy.FireKeyCode:   {506, 76},
			scrcpy.VisionKeyCode: {1600, 600},
			//sdl.K_w:            {361, 699},
			//sdl.K_s:            {357, 924},
			//sdl.K_a:            {245, 811},
			//sdl.K_d:            {470, 808},
			sdl.K_SPACE:  {2087, 755},
			sdl.K_c:      {2050, 967},
			sdl.K_LSHIFT: {1871, 1003},
		},
	}
	log.Println(scrcpy.Main(&option))
}
