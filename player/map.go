package player

import (
	"game/config"
	"game/global"
	"game/lib"
	"game/maps"
	"github.com/liangmanlin/gootp/gutil"
	"github.com/liangmanlin/gootp/kernel"
)

const mapModName = "map"

func init() {
	modRouter[global.PLAYER_MOD_MAP] = &MapHandler
}

var MapLoad = func(player *global.Player) {
	rp := lib.GameDB.SyncSelectRow(player.Context.Call, global.TABLE_ROLE_MAP, player.RoleID, player.RoleID)
	player.Map = rp.(*global.RoleMap)
}

var MapPersistent = func(player *global.Player) {
	lib.GameDB.SyncUpdate(global.TABLE_ROLE_MAP, player.RoleID, kernel.DeepCopy(player.Map))
}

var MapExit = func(player *global.Player) {
	// 无论如何都会在2s后退出
	StartTimer(player, TIMER_EXIT, 0, 2*1000, 1, MapExitTimeOut)
	MapCastMap(player, global.MAP_MOD_ROLE, &global.RoleExitMap{RoleID: player.RoleID})
}

var MapExitTimeOut = func(player *global.Player) {
	MapExitFinal(player, 1, nil)
}

var MapExitFinal = func(player *global.Player, exitType int, exitData *global.MapRoleExit) {
	player.Exit(kernel.ExitReasonNormal)
	if exitType == 1 {
		// 在尝试一次退出场景
		MapCastMap(player, global.MAP_MOD_ROLE, &global.RoleExitMap{RoleID: player.RoleID})
	}else{
		player.Map.X = gutil.Round(exitData.X)
		player.Map.Y = gutil.Round(exitData.Y)
		player.DirtyMod[mapModName] = true
	}
}

// 登录后首次进入地图
var MapFirstEnter = func(player *global.Player) {
	// 判断当前地图是否可以进入
	info := player.Map
	CheckCanEnter(info)
	base := player.Base
	prop := player.Prop
	// 构造PMapRole
	mapRole := &global.PMapRole{
		RoleID: player.RoleID,
		Name: base.Name,
		HeroType: base.HeroType,
		ServerID: lib.GetServerID(),
		Level: int16(base.Level),
		State: maps.ACTOR_STATE_NORMAL,
		Camp: 0,
		Skin: base.Skin,
		Pos: &global.PPos{X: float32(info.X),Y: float32(info.Y),Dir: 270},
		HP: prop.MaxHP,
		MaxHP: prop.MaxHP,
		MoveSpeed: prop.MoveSpeed,
	}
	actor := maps.NewActor(global.ACTOR_ROLE,player.RoleID)
	baseProp := *prop
	actor.BaseProp = &baseProp
	enter := &maps.MapChangeData{
		RoleID: player.RoleID,
		MapInfo: mapRole,
		Actor: actor,
		RoleData: &maps.MapRoleData{
			Player:player.Self(),
			TcpPid: player.GWPid,
			Skills: SkillGetFightSkill(player),
		},
		IsFirstEnter: true,
	}
	mapPid := lib.GetMapPid(info.MapName)
	player.MapPid = mapPid
	lib.CastMap(mapPid, global.MAP_MOD_ROLE, enter)
}

func MapCastMap(player *global.Player, mod int32, msg interface{}) {
	lib.CastMap(player.MapPid, mod, msg)
}

func CheckCanEnter(info *global.RoleMap)  {
	if lib.GetMapPid(info.MapName)== nil {
		// 暂时先简单处理回主城
		info.MapID = int32(101)
		info.MapName = lib.NormalMapName(info.MapID)
		cfg := config.Maps.Get(info.MapID)
		info.X = cfg.BornX
		info.Y = cfg.BornY
		info.Node = kernel.SelfNode().Name()
	}
}

var MapHandler MsgHandler = func(player *global.Player, msg interface{}) {
	switch m := msg.(type) {
	case *global.MapRoleEnter:
		MapRoleEnter(player,m)
	case *global.MapRoleExit:
		MapExitFinal(player, 0,m)
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