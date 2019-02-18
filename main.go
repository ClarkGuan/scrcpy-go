package main

import (
	"flag"
	"log"
	"time"

	"github.com/ClarkGuan/go-sdl2/sdl"
	"github.com/ClarkGuan/scrcpy-go/scrcpy"
)

func main() {
	log.Printf("SDL %d.%d.%d\n", sdl.MAJOR_VERSION, sdl.MINOR_VERSION, sdl.PATCHLEVEL)

	var debugLevel int
	flag.IntVar(&debugLevel, "log", 0, "日志等级设置")
	flag.Parse()

	option := scrcpy.Option{
		Debug:   scrcpy.DebugLevelWrap(debugLevel),
		BitRate: 8000000,
		MaxSize: 0,
		Port:    27183,
		KeyMap: map[int]scrcpy.UserOperation{
			// 开火键
			scrcpy.FireKeyCode: &scrcpy.Point{416, 86},
			// 视野中心坐标
			scrcpy.VisionKeyCode: &scrcpy.Point{1525, 545},
			// 方向键 前
			scrcpy.FrontKeyCode: &scrcpy.Point{346, 689},
			// 方向键 后
			scrcpy.BackKeyCode: &scrcpy.Point{346, 913},
			// 跳/紧急停车
			sdl.K_SPACE: &scrcpy.Point{1883, 564},
			// 趴/下车
			sdl.K_c: &scrcpy.Point{1877, 413},
			// 蹲/加速/下沉
			sdl.K_LSHIFT: &scrcpy.Point{1716, 817},
			// 换弹/投掷距离切换
			sdl.K_r: &scrcpy.Point{1623, 1013},
			// 准镜/喇叭
			sdl.K_e: &scrcpy.Point{1995, 730},
			// 左摆头
			sdl.K_q: &scrcpy.Point{352, 395},
			// 救人/上浮
			sdl.K_z: &scrcpy.Point{1718, 638},
			// 舔包
			sdl.K_t: &scrcpy.SPoint{1444, 274},
			// 打开/收起拾取列表
			sdl.K_y: &scrcpy.Point{1520, 281},
			// 拾取物品1
			sdl.K_f: &scrcpy.Point{1447, 377},
			// 拾取物品2
			sdl.K_g: &scrcpy.Point{1447, 490},
			// 拾取物品3
			sdl.K_h: &scrcpy.Point{1447, 599},
			// 拾取物品4
			sdl.K_j: &scrcpy.Point{1395, 670},
			// 开/关门
			sdl.K_v: &scrcpy.Point{1424, 745},
			// 1号武器
			sdl.K_1: &scrcpy.Point{967, 983},
			// 2号武器
			sdl.K_2: &scrcpy.Point{1205, 977},
			// 使用医疗物品
			sdl.K_3: &scrcpy.Point{715, 1013},
			// 使用投掷物品
			sdl.K_4: &scrcpy.Point{1444, 1020},
			// 打开医疗物品列表
			sdl.K_6: &scrcpy.Point{715, 930},
			// 打开投掷物品列表
			sdl.K_7: &scrcpy.Point{1441, 930},
			// 1号武器单发
			sdl.K_b: &scrcpy.Point{950, 907},
			// 2号武器单发
			sdl.K_n: &scrcpy.Point{1162, 904},
			// 3号武器
			sdl.K_8: &scrcpy.Point{1298, 907},
			// 背包列表
			sdl.K_TAB: &scrcpy.SPoint{76, 1003},
			// 地图
			sdl.K_m: &scrcpy.SPoint{2020, 53},
			// 打开准镜列表
			sdl.K_x: &scrcpy.Point{2014, 450},
			// 比例尺放大
			sdl.K_COMMA: &scrcpy.Point{2008, 264},
			// 比例尺缩小
			sdl.K_PERIOD: &scrcpy.Point{2018, 825},
			// 人物置中
			sdl.K_SLASH: &scrcpy.Point{1885, 1021},
			// 取消标记点
			sdl.K_QUOTE: &scrcpy.Point{1477, 1015},
			// 取消投掷
			sdl.K_BACKQUOTE: &scrcpy.Point{662, 562},
			// 准镜比例缩小
			sdl.K_k: []*scrcpy.PointMacro{{scrcpy.Point{680, 217}, 100 * time.Millisecond},
				{scrcpy.Point{680, 601}, 30 * time.Millisecond},
				{scrcpy.Point{674, 223}, 0}},
			// 准镜比例放大
			sdl.K_l: []*scrcpy.PointMacro{{scrcpy.Point{678, 217}, 100 * time.Millisecond},
				{scrcpy.Point{680, 327}, 30 * time.Millisecond},
				{scrcpy.Point{678, 217}, 0}},
			// 语音：前方有敌人
			sdl.K_u: []*scrcpy.PointMacro{{scrcpy.Point{2010, 358}, 100 * time.Millisecond},
				{scrcpy.Point{1720, 156}, 0}},
			// 语音：我这有物资
			sdl.K_i: []*scrcpy.PointMacro{{scrcpy.Point{2010, 358}, 100 * time.Millisecond},
				{scrcpy.Point{1738, 233}, 0}},
			// 打开团队语音（发出声音）
			sdl.K_o: []*scrcpy.PointMacro{{scrcpy.Point{1748, 197}, 100 * time.Millisecond},
				{scrcpy.Point{1577, 164}, 0}},
			// 关闭团队语音（发出声音）
			sdl.K_p: []*scrcpy.PointMacro{{scrcpy.Point{1752, 203}, 100 * time.Millisecond},
				{scrcpy.Point{1654, 162}, 0}},
		},
		CtrlKeyMap: map[int]scrcpy.UserOperation{
			// 准镜切换1
			sdl.K_1: &scrcpy.Point{1794, 457},
			// 准镜切换2
			sdl.K_2: &scrcpy.Point{1868, 460},
			// 准镜切换3
			sdl.K_3: &scrcpy.Point{1755, 576},
			// 准镜切换4
			sdl.K_4: &scrcpy.Point{1855, 573},
			// 准镜切换5
			sdl.K_5: &scrcpy.Point{1759, 695},
			// 准镜切换6
			sdl.K_6: &scrcpy.Point{1878, 692},
			// 准镜切换7
			sdl.K_7: &scrcpy.Point{1755, 811},
			// 打开团队语音（接收声音）
			sdl.K_o: []*scrcpy.PointMacro{{scrcpy.Point{1748, 115}, 100 * time.Millisecond},
				{scrcpy.Point{1546, 133}, 0}},
			// 关闭团队语音（接收声音）
			sdl.K_p: []*scrcpy.PointMacro{{scrcpy.Point{1755, 123}, 100 * time.Millisecond},
				{scrcpy.Point{1616, 131}, 0}},
		},
		MouseKeyMap: map[uint8]scrcpy.UserOperation{
			// 右摆头
			sdl.BUTTON_RIGHT: &scrcpy.Point{507, 399},
		},
	}
	log.Println(scrcpy.Main(&option))
}
