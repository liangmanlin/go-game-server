package player

import (
	"game/global"
	"github.com/liangmanlin/gootp/kernel"
)

func init() {
	modRouter[global.PLAYER_MOD_ROLE] = &RoleHandler
}

var RoleHandler MsgHandler = func(ctx *kernel.Context, player *global.Player, msg interface{}) {
	switch m := msg.(type) {
	case *global.RoleDeadArg:
		RoleDead(player,m)
	default:
		kernel.UnHandleMsg(msg)
	}
}

var RoleDead = func(player *global.Player,arg *global.RoleDeadArg) {

}


