package maps

type skillCallBackFunc = func(state *MapState,entity *SkillEntity,targets []MapInfo,args []int32)

var SkillCallBackMap = make(map[string]*skillCallBackFunc,100)

var AddSelfBuff skillCallBackFunc = func(state *MapState, entity *SkillEntity, targets []MapInfo,args []int32) {

}

func init()  {
	SkillCallBackMap[`add_self_buff`] = &AddSelfBuff
}