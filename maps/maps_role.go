package maps

import (
	"game/global"
	"game/lib"
	"github.com/liangmanlin/gootp/gutil"
	"github.com/liangmanlin/gootp/kernel"
)

const (
	P_MAP_ROLE_HP int32 = 11
	P_MAP_ROLE_MAXHP int32 = 12
	P_MAP_ROLE_MOVESPEED int32 = 13
)

var RoleSecondLoop = func(state *MapState, now2 int64) {

}

var RoleReduceHP = func(state *MapState, srcActor, destActor *MapActor, destMI *global.PMapRole, damage int32) {
	if destMI.HP > damage {
		destMI.HP -= damage
	} else {
		RoleDead(state, srcActor, destActor, destMI)
		destMI.HP = 0
	}
}

var RoleDead = func(state *MapState, srcActor, destActor *MapActor, destMI *global.PMapRole) {
	roleID := destMI.RoleID
	destActor.State = ACTOR_STATE_DEAD
	destMI.State = ACTOR_STATE_DEAD
	RoleStopMove(state, roleID,true)
	pos := destMI.Pos
	// 广播通知死亡
	proto := &global.MapTocActorDead{ActorID: roleID, ActorType: global.ACTOR_ROLE}
	state.BroadCastPos(pos, proto)
	// 通知player
	arg := &global.RoleDeadArg{SrcID: srcActor.ActorID, SrcType: srcActor.ActorType}
	state.CastPlayer(roleID, global.PLAYER_MOD_ROLE,arg)
}

func RoleStopMove(state *MapState, roleID int64,broadCast bool) {
	aKey := AKey{ActorType: global.ACTOR_ROLE, ActorID: roleID}
	actor := state.GetActor(aKey)
	if actor.IsMove {
		DelMove(state.MoveActor, aKey)
		actor.IsMove = false
		mapInfo := state.MapRoles[roleID]
		mapInfo.MovePath = nil
		if broadCast {
			pos := mapInfo.Pos
			proto := &global.MapTocMoveStop{ActorType: global.ACTOR_ROLE, ActorID: roleID, X: pos.X, Y: pos.Y, Dir: pos.Dir}
			state.BroadCastPosExclude(mapInfo.Pos, roleID, proto)
		}
	}
}

func RoleUpdateHPMoveSpeed(state *MapState, actor *MapActor, upList []global.PKV) {
	MI := state.MapRoles[actor.ActorID]
	ul := make([]*global.PKV,0,len(upList))
	for i := range upList {
		v := &upList[i]
		switch v.Key {
		case lib.PROP_MaxHP:
			if MI.HP >= v.Value {
				MI.HP = v.Value
			}else{
				MI.HP = gutil.Trunc(float32(MI.HP)/float32(MI.MaxHP)*float32(v.Value))
			}
			MI.MaxHP = v.Value
			v.Key = P_MAP_ROLE_MAXHP
			ul = append(ul,v,&global.PKV{Key: P_MAP_ROLE_HP,Value: MI.HP})
		case lib.PROP_MoveSpeed:
			if v.Value != MI.MoveSpeed {
				MI.MoveSpeed = v.Value
				v.Key = P_MAP_ROLE_MOVESPEED
				ul = append(ul,v)
			}
		}
	}
	state.UpdateRoleInfo(MI.RoleID,MI.Pos,ul)
}

var MapRoleHandle MsgHandler = func(ctx *kernel.Context, state *MapState, msg interface{}) {
	switch m := msg.(type) {
	case *global.RoleUpdateProps:
		MapRoleUpdateProps(state,m)
	case *global.RoleExitMap:
		// 包装成结构，方便扩展
		MapRoleChangeMap(ctx,state,ChangeMapArg{
		RoleID: m.RoleID,
		DestMapID: state.Config.MapID,
		DestMapName: state.Name,
		DestNode: kernel.SelfNode(),
		IsExit: true,
		})
	default:
		kernel.UnHandleMsg(msg)
	}
}

var MapRoleChangeMap = func(ctx *kernel.Context, state *MapState,arg ChangeMapArg) {
	roleID := arg.RoleID
	// 在真正退出之前执行回调
	state.Mod.RoleLeave(state,ctx,roleID,arg.IsExit)
	// 退出aoi
	mi := state.MapRoles[roleID]
	area := GetAreaByPos(mi.Pos)
	state.LeaveActor(area,global.ACTOR_ROLE,roleID)
	RoleStopMove(state,roleID,false)
	proto := &global.MapTocActorLeaveArea{ActorType: global.ACTOR_ROLE,ActorID: roleID}
	state.BroadCastAreas(Get9AreaByArea(area),proto)
	if arg.IsExit {
		// 退出地图，通知player下线
		MapRoleExitMap(state,roleID,mi)
		// 清理数据
		MapRoleCleanData(state,roleID,false)
	}else{
		// 构造迁移数据
		changeData := MapRoleCleanData(state,roleID,true)
		ChangeMap(state,arg.DestNode,arg.DestMapName,changeData)
	}
}

var ChangeMap = func(state *MapState,destNode *kernel.Node,mapName string,changeData *MapChangeData) {
	if destNode.Equal(kernel.SelfNode()){
		// 相同节点
		if pid := lib.GetMapPid(mapName);pid != nil{
			lib.CastMap(pid,global.MAP_MOD_ROLE,changeData)
		}else{
			// 回出生点
			lib.CastPlayer(changeData.RoleData.Player,global.PLAYER_MOD_MAP,changeData)
		}
	}else{
		// 跨节点，目前只支持集群模式
		if kernel.IsNodeConnect(destNode.Name()) {
			kernel.CastNameNode("map_agent",destNode,&AgentChangeMap{MapName: mapName,Change: changeData})
		}else{
			// 回出生点
			lib.CastPlayer(changeData.RoleData.Player,global.PLAYER_MOD_MAP,changeData)
		}
	}
}

var MapRoleExitMap = func(state *MapState,roleID int64,mi *global.PMapRole) {
	pos := mi.Pos
	data := &global.MapRoleExit{X: pos.X,Y:pos.Y}
	state.CastPlayer(roleID,global.PLAYER_MOD_MAP,data)
}

var MapRoleCleanData = func(state *MapState,roleID int64,rt bool)(rs *MapChangeData) {
	if rt {
		rs = &MapChangeData{
			RoleID: roleID,
			MapInfo: state.MapRoles[roleID],
			Actor: state.GetActorByID(global.ACTOR_ROLE,roleID),
			RoleData: state.Roles[roleID],
		}
	}
	delete(state.MapRoles,roleID)
	delete(state.Actors,AKey{roleID,global.ACTOR_ROLE})
	delete(state.Roles,roleID)
	return
}

var MapRoleUpdateProps = func(state *MapState,msg *global.RoleUpdateProps) {
	actor := state.GetActorByID(global.ACTOR_ROLE,msg.RoleID)
	baseProp := actor.BaseProp
	keys := make([]int32,0,len(msg.UP))
	for _,v :=range msg.UP{
		lib.SetPropValue(baseProp,v.Key,v.Value)
		keys = append(keys,v.Key)
	}
	PropCalcProp(state,actor,keys)
}

func init() {
	modRouter[global.MAP_MOD_ROLE] = &MapRoleHandle
}