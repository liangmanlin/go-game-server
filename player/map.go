package player

import (
	"game/config"
	"game/global"
	"game/lib"
	"game/maps"
	"github.com/liangmanlin/gootp/db"
	"github.com/liangmanlin/gootp/gutil"
	"github.com/liangmanlin/gootp/kernel"
)

const mapModName = "map"

func init() {
	modRouter[global.PLAYER_MOD_MAP] = &MapHandler
}

var MapLoad = func(ctx *kernel.Context, player *global.Player) {
	rp := db.SyncSelectRow(ctx, global.TABLE_ROLE_MAP, player.RoleID, player.RoleID)
	player.Map = rp.(*global.RoleMap)
}

var MapPersistent = func(ctx *kernel.Context, player *global.Player) {
	db.SyncUpdate(global.TABLE_ROLE_MAP, player.RoleID, kernel.DeepCopy(player.Map))
}

var MapExit = func(player *global.Player, ctx *kernel.Context) {
	// 无论如何都会在2s后退出
	StartTimer(player, TIMER_EXIT, 0, 2*1000, 1, MapExitTimeOut, ctx)
	MapCastMap(player, global.MAP_MOD_ROLE, &global.RoleExitMap{RoleID: player.RoleID})
}

var MapExitTimeOut = func(player *global.Player, ctx *kernel.Context) {
	MapExitFinal(player, 1, ctx, nil)
}

var MapExitFinal = func(player *global.Player, exitType int, ctx *kernel.Context, exitData *global.MapRoleExit) {
	ctx.Exit(kernel.ExitReasonNormal)
	if exitType == 1 {
		// 在尝试一次退出场景
		MapCastMap(player, global.MAP_MOD_ROLE, &global.RoleExitMap{RoleID: player.RoleID})
	}else{
		player.Map.X = gutil.Round(exitData.X)
		player.Map.Y = gutil.Round(exitData.Y)
		player.DirtyMod[mapModName] = true
	}
}

func MapCastMap(player *global.Player, mod int32, msg interface{}) {
	lib.CastMap(player.MapPid, mod, msg)
}

var MapHandler MsgHandler = func(ctx *kernel.Context, player *global.Player, msg interface{}) {
	switch m := msg.(type) {
	case *global.MapRoleEnter:
		MapRoleEnter(player,m)
	case *global.MapRoleExit:
		MapExitFinal(player, 0, ctx,m)
	case *maps.MapChangeData:
		// 需要回到主城
		mainMapID := int32(101)
		mapName := lib.NormalMapName(mainMapID)
		cfg := config.Maps.Get(mainMapID)
		m.MapInfo.Pos.X = float32(cfg.BornX)
		m.MapInfo.Pos.Y = float32(cfg.BornY)
		pid := lib.GetMapPid(mapName)
		player.MapPid = pid
		MapCastMap(player,global.MAP_MOD_ROLE,m)
	default:
		kernel.UnHandleMsg(msg)
	}
}

var MapRoleEnter = func(player *global.Player,m *global.MapRoleEnter){
	player.MapPid = m.Pid
	player.Map.X = int32(m.X)
	player.Map.Y = int32(m.Y)
	player.Map.MapName = m.MapName
	player.Map.MapID = m.MapID
	player.Map.Node = m.Pid.Node().Name()
}