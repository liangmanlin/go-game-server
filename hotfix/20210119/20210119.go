package main

import (
	"game/global"
	"game/player"
	"github.com/liangmanlin/gootp/kernel"
	"sync/atomic"
	"unsafe"
)

func HotFix(){
	// 这里赋予新的函数
	f := func(player *global.Player, proto interface{}) {
		kernel.ErrorLog("this is new api")
	}
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&player.LoginLogin)),*(*unsafe.Pointer)(unsafe.Pointer(&f)))
}