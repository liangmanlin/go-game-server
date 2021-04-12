package player

import (
	"game/global"
	"github.com/liangmanlin/gootp/kernel"
)

type mod struct {
	name       string
	load       *func(ctx *kernel.Context, player *global.Player)
	persistent *func(ctx *kernel.Context, player *global.Player)
}

type TResult struct {
	OK     bool
	Result interface{}
}
