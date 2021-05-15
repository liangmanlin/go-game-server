package maps

import "game/lib"

func PropRMBuffProp(state *MapState, actor *MapActor, buffID int16) {
	pKey := lib.PropKey{ID: lib.PROP_KEY_BUFF, SubID: int32(buffID)}
	PropRmProp(state, actor, pKey)
}

func PropRmProp(state *MapState, actor *MapActor, pkey lib.PropKey) {
	keys := actor.PropData.RmPropKvs(pkey)
	PropCalcProp(state, actor, keys)
}

var PropCalcProp = func(state *MapState, actor *MapActor, keys []int32) {
	upList := actor.PropData.CalcMapProps(keys, actor.BaseProp)
	totalProp := actor.TotalProp
	var idx int
	for _,v := range upList{
		if v.Key == lib.PROP_MaxHP || v.Key == lib.PROP_MoveSpeed{
			upList[idx] = v
			idx++
		}
		lib.SetPropValue(totalProp,v.Key,v.Value)
	}
	if idx > 0 {
		UpdateHPMoveSpeed(state,actor,upList[0:idx])
	}
}
