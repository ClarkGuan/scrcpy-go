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
			scrcpy.VisionKeyCode: {1525, 545},
			scrcpy.FrontKeyCode:  {350, 695},
			scrcpy.BackKeyCode:   {350, 921},
			sdl.K_SPACE:          {1994, 745},
			sdl.K_c:              {1964, 978},
			sdl.K_LSHIFT:         {1775, 1000},
			sdl.K_r:              {1623, 1013},
			sdl.K_e:              {2000, 566},
			sdl.K_z:              {1080, 662},
			sdl.K_t:              {1444, 274},
			sdl.K_y:              {1520, 281},
			sdl.K_f:              {1447, 377},
			sdl.K_g:              {1447, 490},
			sdl.K_h:              {1447, 599},
			sdl.K_j:              {1447, 689},
			sdl.K_v:              {1424, 745},
			sdl.K_1:              {967, 983},
			sdl.K_2:              {1205, 977},
			sdl.K_3:              {715, 1013},
			sdl.K_4:              {1444, 1020},
			sdl.K_5:              {1911, 420},
			sdl.K_6:              {715, 930},
			sdl.K_7:              {1441, 930},
			sdl.K_b:              {950, 907},
			sdl.K_n:              {1162, 904},
			sdl.K_8:              {1298, 907},
			sdl.K_TAB:            {76, 1003},
			sdl.K_m:              {2020, 53},
			sdl.K_DOWN:           {1706, 838},
			sdl.K_p:              {1901, 586},
		},
	}
	log.Println(scrcpy.Main(&option))
}
