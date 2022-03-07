package player

import (
	"game/config"
	"game/global"
	"game/lib"
)

const propName = "prop"

var PropLoad = func(player *global.Player) {
	rp := lib.GameDB.SyncSelectRow(player.Context.Call, global.TABLE_ROLE_PROP, player.RoleID, player.RoleID)
	player.Prop = rp.(*global.RoleProp).Prop
}

var PropPersistent = func(player *global.Player) {
	p := *player.Prop
	lib.GameDB.SyncUpdate(global.TABLE_ROLE_PROP, player.RoleID, &global.RoleProp{RoleID: player.RoleID, Prop: &p})
}

// 登录之后，从新计算一次属性
var InitProps = func(player *global.Player) {
	pd := propData(player)
	pl := pd.CalcProps(lib.AllKeys())
	prop := ToProp(player.Prop,pl)
	player.Prop = prop
}

func ToProp(prop *global.PProp,pl []*global.PKV) *global.PProp {
	for _,v := range pl {
		lib.SetPropValue(prop,v.Key,v.Value)
	}
	return prop
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
