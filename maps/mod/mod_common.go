package mod

import (
	"game/maps"
	"github.com/liangmanlin/gootp/kernel"
)

// 解决循环依赖
func init() {
	maps.ModMap["mod_common"] = Common
}

var Common = &maps.MapMod{
	Init: func(state *maps.MapState, ctx *kernel.Context, mapID int32, args ...interface{}) {
		maps.CreateMapAllMonster(state)
	},
	RoleEnter: func(state *maps.MapState, ctx *kernel.Context, roleID int64) {

	},
	RoleReconnect: func(state *maps.MapState, ctx *kernel.Context, roleID int64) {

	},
	RoleCanEnter: func(state *maps.MapState, ctx *kernel.Context, roleID int64) bool {
		return true
	},
	RoleLeave: func(state *maps.MapState, ctx *kernel.Context, roleID int64,isExit bool) {

	},
	RoleDead: func(state *maps.MapState, ctx *kernel.Context, roleID int64) {

	},
	RoleRelive: func(state *maps.MapState, ctx *kernel.Context, roleID int64) {

	},
	Handle: func(state *maps.MapState, ctx *kernel.Context, msg interface{}) {

	},
}
