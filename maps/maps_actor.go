package maps

import (
	"game/global"
	"game/lib"
)

func ActorReduceHP(state *MapState, srcActor, destActor *MapActor, target MapInfo, damage int32) {
	if !target.IsAlive() {
		return
	}
	if destActor.ActorType == global.ACTOR_ROLE {
		RoleReduceHP(state, srcActor, destActor, target.(*global.PMapRole), damage)
	} else {
		MonsterReduceHP(state, srcActor, destActor, target.(*global.PMapMonster), damage)
	}
}

func NewActor(actorType int8, actorID int64) *MapActor {
	buffData := &BuffData{
		Type2ID:     make(map[int16]int16),
		Data:        make(map[int16]*BuffInfo),
		EffectCount: make([]int16,2),
	}
	propData := lib.NewPropData()
	return &MapActor{ActorType: actorType, ActorID: actorID, State: ACTOR_STATE_NORMAL,BuffData: buffData,PropData: propData}
}

func (m *MapActor) IsAlive() bool {
	return m.State != ACTOR_STATE_DEAD
}

func UpdateHPMoveSpeed(state *MapState,actor *MapActor,upList []global.PKV)  {
	if actor.ActorType == global.ACTOR_ROLE {
		RoleUpdateHPMoveSpeed(state,actor,upList)
	}else{
		MonsterUpdateHPMoveSpeed(state,actor,upList)
	}
}