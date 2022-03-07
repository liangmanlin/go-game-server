package proto

import (
	"game/global"
)

func Router(router map[int]*global.HandleFunc,protoID int,proto interface{},player *global.Player) bool  {
	if f,ok := router[protoID];ok{
		(*f)(player,proto)
		return true
	}
	return false
}
