package player

import (
	"game/global"
	"github.com/liangmanlin/gootp/kernel"
)

var LoginLogin = func(player *global.Player, proto interface{}) {
	kernel.ErrorLog("%#v", proto)
}
