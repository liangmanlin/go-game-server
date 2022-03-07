package maps

import (
	"game/config"
	"game/global"
)

const (
	BUFF_INV_TYPE_NORMAL int32 = iota + 1
	BUFF_INV_TYPE_INV
)

const (
	BUFF_EFFECT_COUNT_NO_MOVE = iota
	BUFF_EFFECT_COUNT_NO_SKILL
)

func BuffLoop(state *MapState, now2 int64) {
	// 先遍历玩家buff
	roles := state.MapRoles
	for _, role := range roles {
		BuffLoopMapInfo(state, role, now2)
	}
	//monsters := state.MapMonsters
	//for _, monster := range monsters {
	//	if monster.IsAlive() {
	//		BuffLoopMapInfo(state, monster, now2)
	//	}
	//}
}

var BuffLoopMapInfo = func(state *MapState, MI MapInfo, now2 int64) {
	buffs := MI.GetBuffs()
	if len(buffs) == 0 {
		return
	}
	actor := state.GetActorByID(MI.Type(), MI.ID())
	data := actor.BuffData
	var idx int
	for _, buff := range buffs {
		cfg := config.Buffs.Get(int32(buff.ID))
		if cfg.InvType == BUFF_INV_TYPE_INV {
			bd := data.Data[buff.ID]
			if now2 >= bd.TickTime {
				// TODO 执行效果
				bd.TickTime += int64(cfg.Inv)
			}
		}
		if now2 >= buff.EndTime {
			// 删除buff，需要移除一些效果
			DelBuffEffect(state,actor,MI,buff,cfg)
			DelBuffNotice(state, cfg, MI)
		} else {
			buffs[idx] = buff
			idx++
		}
	}
	if idx > 0 {
		MI.SetBuffs(buffs[0:idx])
	} else {
		// gc
		MI.SetBuffs(nil)
	}
}

var DelBuffEffectCount = func(data *BuffData, cfg *config.DefBuffs) {
	if cfg.NoMmove > 0 {
		data.EffectCount[BUFF_EFFECT_COUNT_NO_MOVE]--
	}
	if cfg.NoSkill > 0 {
		data.EffectCount[BUFF_EFFECT_COUNT_NO_SKILL]--
	}
}

var DelBuffNotice = func(state *MapState, cfg *config.DefBuffs, MI MapInfo) {
	switch cfg.NoticeType {
	case 1:
		// 广播
		proto := &global.MapTocDelBuff{ActorType: MI.Type(), ActorID: MI.ID(), BuffID: int16(cfg.Id)}
		state.BroadCastPos(MI.GetPos(), proto)
	case 2:
		// 只通知自己
		if MI.Type() == global.ACTOR_ROLE {
			proto := &global.MapTocDelBuff{ActorType: MI.Type(), ActorID: MI.ID(), BuffID: int16(cfg.Id)}
			state.SendRoleProto(MI.ID(), proto)
		}
	}
}

// buff通常都是有一些效果的，例如加属性
var DelBuffEffect = func(state *MapState,actor *MapActor,MI MapInfo,buff *global.PBuff,cfg *config.DefBuffs) {
	data := actor.BuffData
	// 删除数据
	delete(data.Type2ID, buff.Type)
	delete(data.Data, buff.ID)
	DelBuffEffectCount(data, cfg)
	switch cfg.EffectType {
	case 1:// 加了属性
		PropRMBuffProp(state,actor,buff.ID)
	}
}
