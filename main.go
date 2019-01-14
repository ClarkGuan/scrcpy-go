package main

import (
	"log"

	"github.com/ClarkGuan/scrcpy-go/scrcpy"
)

func main() {
	option := scrcpy.Option{
		Debug:   true,
		BitRate: 8000000,
		MaxSize: 0,
		Port:    27183,
	}
	log.Println(scrcpy.Main(&option))
}
