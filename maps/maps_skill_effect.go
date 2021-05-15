package maps

import (
	"game/global"
	"github.com/liangmanlin/gootp/gutil"
)

const (
	EFFECT_TYPE_DAMAGE int8 = iota + 1
)

type skillEffectFunc = func(state *MapState, entity *SkillEntity, srcActor *MapActor, target MapInfo, allTargets []MapInfo,
	effectList []*global.PSkillEffect, args []int32) []*global.PSkillEffect

var SkillEffectMap = make(map[string]*skillEffectFunc, 100)

var CommonAttack skillEffectFunc = func(state *MapState, entity *SkillEntity, srcActor *MapActor, target MapInfo, allTargets []MapInfo,
	effectList []*global.PSkillEffect, args []int32) []*global.PSkillEffect {
	destType := target.Type()
	destID := target.ID()
	destActor := state.GetActorByID(destType, destID)
	// 先简单计算,理论上应该计算暴击，闪避等
	srcProp := srcActor.TotalProp
	destProp := destActor.TotalProp
	damage := gutil.MaxInt32(srcProp.PhyAttack-destProp.PhyDefence, 1)
	// 扣血
	ActorReduceHP(state,srcActor,destActor,target,damage)
	effect := NewEffect(destType, destID, EFFECT_TYPE_DAMAGE, damage)
	return append(effectList,effect)
}

func init() {
	SkillEffectMap["common_attack"] = &CommonAttack
}
