package main

import (
	"flag"
	"io/ioutil"
	"log"
	"path/filepath"
	"time"

	"github.com/ClarkGuan/go-sdl2/sdl"
	"github.com/ClarkGuan/scrcpy-go/scrcpy"
	"gopkg.in/yaml.v2"
)

type EntryFile struct {
	Entries []*Entry `yaml:"keys"`
}

type Entry struct {
	Code        string        `yaml:"code"`
	Point       *EntryPoint   `yaml:"point"`
	Comment     string        `yaml:"comment"`
	Macro       []*EntryMacro `yaml:"macro"`
	ShowPointer bool          `yaml:"show_pointer"`
	Type        string        `yaml:"type"`
}

type EntryPoint struct {
	X int `yaml:"x"`
	Y int `yaml:"y"`
}

type EntryMacro struct {
	Point *EntryPoint `yaml:"point"`
	Delay int         `yaml:"delay"`
}

func main() {
	log.Printf("SDL %d.%d.%d\n", sdl.MAJOR_VERSION, sdl.MINOR_VERSION, sdl.PATCHLEVEL)

	var debugLevel int
	var bitRate int
	//var maxSize int
	var port int
	var settingFile string
	flag.IntVar(&debugLevel, "log", 0, "日志等级设置")
	flag.IntVar(&bitRate, "bitrate", 8000000, "视频码率")
	//flag.IntVar(&maxSize, "maxsize", 0, "未知")
	flag.IntVar(&port, "port", 27183, "adb 端口号")
	flag.StringVar(&settingFile, "config", filepath.Join(sdl.GetBasePath(), "res", "keys.yml"), "配置文件路径")
	flag.Parse()

	content, err := ioutil.ReadFile(settingFile)
	if err != nil {
		log.Fatalln(err)
	}
	var entryFile EntryFile
	if err = yaml.Unmarshal(content, &entryFile); err != nil {
		log.Fatalln(err)
	}

	keyMap, ctrlKeyMap := make(map[int]scrcpy.UserOperation), make(map[int]scrcpy.UserOperation)
	mouseKeyMap := make(map[uint8]scrcpy.UserOperation)

	for _, entry := range entryFile.Entries {
		switch entry.Type {
		case "":
			if keyCode, ok := scrcpy.KeyCodeConstMap[entry.Code]; ok {
				keyMap[keyCode] = parseUserOperation(entry)
			} else {
				keyCode = int(sdl.GetKeyFromName(entry.Code))
				if keyCode == sdl.K_UNKNOWN {
					log.Fatalln("unknown key code:", entry.Code)
				}
				keyMap[keyCode] = parseUserOperation(entry)
			}

		case "ctrl":
			if keyCode, ok := scrcpy.KeyCodeConstMap[entry.Code]; ok {
				ctrlKeyMap[keyCode] = parseUserOperation(entry)
			} else {
				keyCode = int(sdl.GetKeyFromName(entry.Code))
				if keyCode == sdl.K_UNKNOWN {
					log.Fatalln("unknown key code:", entry.Code)
				}
				ctrlKeyMap[keyCode] = parseUserOperation(entry)
			}

		case "mouse":
			if keyCode, ok := scrcpy.MouseButtonMap[entry.Code]; ok {
				mouseKeyMap[keyCode] = parseUserOperation(entry)
			} else {
				log.Fatalln("unknown mouse code:", entry.Code)
			}

		default:
			panic("can't reach here")
		}
	}

	option := scrcpy.Option{
		Debug:       scrcpy.DebugLevelWrap(debugLevel),
		BitRate:     bitRate,
		MaxSize:     0,
		Port:        port,
		KeyMap:      keyMap,
		CtrlKeyMap:  ctrlKeyMap,
		MouseKeyMap: mouseKeyMap,
	}
	log.Println(scrcpy.Main(&option))
}

func parseUserOperation(entry *Entry) scrcpy.UserOperation {
	if entry.Point != nil {
		if entry.ShowPointer {
			return &scrcpy.SPoint{uint16(entry.Point.X), uint16(entry.Point.Y)}
		} else {
			return &scrcpy.Point{uint16(entry.Point.X), uint16(entry.Point.Y)}
		}
	} else if len(entry.Macro) > 0 {
		var list []*scrcpy.PointMacro
		for _, m := range entry.Macro {
			list = append(list, &scrcpy.PointMacro{
				Point:    scrcpy.Point{uint16(m.Point.X), uint16(m.Point.Y)},
				Interval: time.Duration(m.Delay) * time.Millisecond})
		}
		return list
	} else {
		panic("can't reach here")
	}
}
