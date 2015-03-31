package main

import (
	"fmt"
	"github.com/xeniumd-china/magpie/global"
	"strings"
)

func main() {
	fmt.Println(global.GetFirstLocalIP())
	s := strings.SplitN("a=th=45", "=", 2)
	fmt.Println(s)
	//	macs, _ := global.GetLocalMac()
	//	for _, mac := range macs {
	//		fmt.Println(mac)
	//	}

}
