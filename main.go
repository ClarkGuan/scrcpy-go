package main

import (
	"log"

	"github.com/ClarkGuan/go-sdl2/sdl"
	"github.com/ClarkGuan/scrcpy-go/scrcpy"
)

func main() {
	log.Printf("SDL %d.%d.%d\n", sdl.MAJOR_VERSION, sdl.MINOR_VERSION, sdl.PATCHLEVEL)
	option := scrcpy.Option{
		Debug:   scrcpy.DebugLevelMin,
		BitRate: 8000000,
		MaxSize: 0,
		Port:    27183,
		KeyMap: map[int]*scrcpy.Point{
			scrcpy.FireKeyCode:   {416, 86},
			scrcpy.VisionKeyCode: {1525, 545},
			scrcpy.FrontKeyCode:  {350, 695},
			scrcpy.BackKeyCode:   {350, 921},
			sdl.K_SPACE:          {1994, 745},
			sdl.K_c:              {1964, 978},
			sdl.K_LSHIFT:         {1775, 1000},
			sdl.K_r:              {1623, 1013},
			sdl.K_e:              {2000, 566},
			sdl.K_q:              {1755, 291},
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
			sdl.K_UP:             {1755, 642},
			sdl.K_DOWN:           {1706, 838},
			sdl.K_p:              {1901, 586},
			sdl.K_x:              {2014, 450},
		},
		CtrlKeyMap: map[int]*scrcpy.Point{
			sdl.K_1: {1794, 457},
			sdl.K_2: {1868, 460},
			sdl.K_3: {1755, 576},
			sdl.K_4: {1855, 573},
			sdl.K_5: {1759, 695},
			sdl.K_6: {1878, 692},
			sdl.K_7: {1755, 811},
		},
	}
	log.Println(scrcpy.Main(&option))
}
