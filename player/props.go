package player

import (
	"game/config"
	"game/global"
	"game/lib"
	"github.com/liangmanlin/gootp/db"
	"github.com/liangmanlin/gootp/kernel"
)

const propName = "prop"

var PropLoad = func(ctx *kernel.Context, player *global.Player) {
	rp := db.SyncSelectRow(ctx, global.TABLE_ROLE_PROP, player.RoleID, player.RoleID)
	player.Prop = rp.(*global.RoleProp).Prop
}

var PropPersistent = func(ctx *kernel.Context, player *global.Player) {
	p := *player.Prop
	db.SyncUpdate(global.TABLE_ROLE_PROP, player.RoleID, &global.RoleProp{RoleID: player.RoleID, Prop: &p})
}

// 添加属性
var PropSetKvs = func(player *global.Player, key lib.PropKey, props []config.KV) {
	pd := propData(player)
	keys := pd.SetPropKvs(key, props)
	PropCalcTotal(player, pd, keys)
}

// 删除属性
var PropRmProp = func(player *global.Player, key lib.PropKey) {
	pd := propData(player)
	keys := pd.RmPropKvs(key)
	PropCalcTotal(player, pd, keys)
}

// 从新计算属性，以及战力
var PropCalcTotal = func(player *global.Player, pd *lib.PropData, keys []int32) {
	kvList := pd.CalcProps(keys)
	prop := player.Prop
	// 更新属性
	for _, v := range kvList {
		lib.SetPropValue(prop, v.Key, v.Value)
	}
	if len(kvList) > 0 {
		// 发送到场景
		MapCastMap(player, global.MAP_MOD_ROLE, &global.RoleUpdateProps{RoleID: player.RoleID, UP: kvList})
		// 通知客户端
		SendProto(player, &global.RoleTocUPProps{UP: kvList})
	}
}

func propData(player *global.Player) *lib.PropData {
	return (*lib.PropData)(player.PropData)
}
