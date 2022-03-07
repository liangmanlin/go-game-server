package player

import (
	"game/global"
	"game/lib"
)

const attrName = "attr"

const (
	backup_diamond int32 = iota
	backup_gold
)

var AttrLoad = func(player *global.Player) {
	attr := lib.GameDB.SyncSelectRow(player.Context.Call, global.TABLE_ROLE_ATTR, player.RoleID, player.RoleID)
	player.Attr = attr.(*global.PRoleAttr)
}

var AttrPersistent = func(player *global.Player) {
	a := *player.Attr // 没有二层数据，可以直接拷贝，比起使用反射，效率极高
	lib.GameDB.SyncUpdate(global.TABLE_ROLE_ATTR, player.RoleID, &a)
}

var AttrReduceDiamond = func(player *global.Player, reduce int32) {
	checkTransaction(player)
	rd := int64(reduce)
	if player.Attr.Diamond < rd {
		Abort(&global.PMsg{})
	}
	Backup(player, global.BackupKey{BackupID: global.BACKUP_ATTR, ID: backup_diamond}, "Attr.Diamond", player.Attr.Diamond)
	player.Attr.Diamond -= rd
	AddDBQueue(player,AttrDBOP,global.DB_OP_UPDATE,nil)
}

var AttrAddDiamond = func(player *global.Player, add int32) {
	checkTransaction(player)
	Backup(player, global.BackupKey{BackupID: global.BACKUP_ATTR, ID: backup_diamond}, "Attr.Diamond", player.Attr.Diamond)
	player.Attr.Diamond += int64(add)
	AddDBQueue(player,AttrDBOP,global.DB_OP_UPDATE,nil)
}

var AttrReduceGold = func(player *global.Player, reduce int32) {
	checkTransaction(player)
	rd := int64(reduce)
	if player.Attr.Gold < rd {
		Abort(&global.PMsg{})
	}
	Backup(player, global.BackupKey{BackupID: global.BACKUP_ATTR, ID: backup_gold}, "Attr.Gold", player.Attr.Gold)
	player.Attr.Gold -= rd
	AddDBQueue(player,AttrDBOP,global.DB_OP_UPDATE,nil)
}

var AttrAddGold = func(player *global.Player, add int32) {
	checkTransaction(player)
	Backup(player, global.BackupKey{BackupID: global.BACKUP_ATTR, ID: backup_gold}, "Attr.Gold", player.Attr.Gold)
	player.Attr.Gold += int64(add)
	AddDBQueue(player,AttrDBOP,global.DB_OP_UPDATE,nil)
}

var AttrDBOP = func(player *global.Player, _ int32, _ interface{}){
	player.DirtyMod[attrName] = true
}