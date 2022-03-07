package player

import (
	"game/global"
)

type mod struct {
	name       string
	load       *func(player *global.Player)
	persistent *func(player *global.Player)
}

type TResult struct {
	OK     bool
	Result interface{}
}

type MsgHandler func(player *global.Player,msg interface{})