package main

import (
	"log"
	"time"

	"github.com/ClarkGuan/go-sdl2/sdl"
	"github.com/ClarkGuan/scrcpy-go/scrcpy"
)

func main() {
	log.Printf("SDL %d.%d.%d\n", sdl.MAJOR_VERSION, sdl.MINOR_VERSION, sdl.PATCHLEVEL)
	option := scrcpy.Option{
		Debug:   scrcpy.DebugLevelDebug,
		BitRate: 8000000,
		MaxSize: 0,
		Port:    27183,
		KeyMap: map[int]*scrcpy.Point{
			// 开火键
			scrcpy.FireKeyCode: {416, 86},
			// 视野中心坐标
			scrcpy.VisionKeyCode: {1525, 545},
			// 方向键 前
			scrcpy.FrontKeyCode: {346, 689},
			// 方向键 后
			scrcpy.BackKeyCode: {346, 913},
			// 跳
			sdl.K_SPACE: {1994, 745},
			// 趴
			sdl.K_c: {1964, 978},
			// 蹲
			sdl.K_LSHIFT: {1775, 1000},
			// 换弹
			sdl.K_r: {1623, 1013},
			// 准镜
			sdl.K_e: {2000, 566},
			// 左摆头
			sdl.K_q: {352, 395},
			// 救人
			sdl.K_z: {1080, 662},
			// 舔包
			sdl.K_t: {1444, 274},
			// 打开/收起拾取列表
			sdl.K_y: {1520, 281},
			// 拾取物品1
			sdl.K_f: {1447, 377},
			// 拾取物品2
			sdl.K_g: {1447, 490},
			// 拾取物品3
			sdl.K_h: {1447, 599},
			// 拾取物品4
			sdl.K_j: {1395, 670},
			// 开/关门
			sdl.K_v: {1424, 745},
			// 1号武器
			sdl.K_1: {967, 983},
			// 2号武器
			sdl.K_2: {1205, 977},
			// 使用医疗物品
			sdl.K_3: {715, 1013},
			// 使用投掷物品
			sdl.K_4: {1444, 1020},
			// 下车
			sdl.K_5: {1911, 420},
			// 打开医疗物品列表
			sdl.K_6: {715, 930},
			// 打开投掷物品列表
			sdl.K_7: {1441, 930},
			// 1号武器单发
			sdl.K_b: {950, 907},
			// 2号武器单发
			sdl.K_n: {1162, 904},
			// 3号武器
			sdl.K_8: {1298, 907},
			// 背包列表
			sdl.K_TAB: {76, 1003},
			// 地图
			sdl.K_m: {2020, 53},
			// 上浮
			sdl.K_UP: {1755, 642},
			// 下沉/氮气加速
			sdl.K_DOWN: {1706, 838},
			// 紧急停车
			sdl.K_p: {1901, 586},
			// 打开准镜列表
			sdl.K_x:     {2014, 450},
			sdl.K_COMMA: {1469, 744},
			// 比例尺放大
			sdl.K_LEFTBRACKET: {2008, 264},
			// 比例尺缩小
			sdl.K_RIGHTBRACKET: {2018, 825},
			// 人物置中
			sdl.K_BACKSLASH: {1885, 1021},
			// 取消标记点
			sdl.K_QUOTE: {1477, 1015},
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
		MouseKeyMap: map[uint8]*scrcpy.Point{
			// 右摆头
			sdl.BUTTON_RIGHT: {507, 399},
			sdl.BUTTON_X1:    {1755, 291},
		},
		PointIntervalKeyMap: map[int][]*scrcpy.PointInterval{
			// 准镜比例缩小
			sdl.K_k: {{scrcpy.Point{680, 217}, 100 * time.Millisecond},
				{scrcpy.Point{680, 601}, 30 * time.Millisecond},
				{scrcpy.Point{674, 223}, 0}},
			// 准镜比例放大
			sdl.K_l: {{scrcpy.Point{678, 217}, 100 * time.Millisecond},
				{scrcpy.Point{680, 327}, 30 * time.Millisecond},
				{scrcpy.Point{678, 217}, 0}},
			// 前方有敌人
			sdl.K_u: {{scrcpy.Point{2010, 358}, 100 * time.Millisecond},
				{scrcpy.Point{1720, 156}, 0}},
			// 我这有物资
			sdl.K_i: {{scrcpy.Point{2010, 358}, 100 * time.Millisecond},
				{scrcpy.Point{1738, 233}, 0}},
		},
	}
	log.Println(scrcpy.Main(&option))
}
