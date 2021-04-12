package proto

import (
	"game/global"
	"github.com/liangmanlin/gootp/kernel"
)

func Router(router map[int]*global.HandleFunc,protoID int,proto interface{},ctx *kernel.Context,player *global.Player) bool  {
	if f,ok := router[protoID];ok{
		(*f)(ctx,player,proto)
		return true
	}
	return false
}
