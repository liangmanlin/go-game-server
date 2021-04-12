package player

import (
	"game/global"
	"github.com/liangmanlin/gootp/kernel"
)

var LoginLogin = func(ctx *kernel.Context,player *global.Player, proto interface{}) {
	kernel.ErrorLog("%#v", proto)
}
