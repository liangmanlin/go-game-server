package lib

import (
	"game/config"
	"game/global"
	"github.com/liangmanlin/gootp/gutil"
	"unsafe"
)

const propSize = int32(unsafe.Sizeof(global.PProp{}) / 4)

var allKeys []int32

func init() {
	allKeys = make([]int32, 0, propSize)
	for i := int32(1); i <= propSize; i++ {
		allKeys = append(allKeys,i)
	}
}

func AllKeys() []int32 {
	return allKeys
}

/*
	属性计算模块
*/

func NewPropData() *PropData {
	return &PropData{
		Cache:   make(map[PropKey][]int32),
		PropMap: make(map[int32]*PropSum),
	}
}

// 添加属性，会替换掉原有的属性
func (p *PropData) SetPropKvs(key PropKey, list []config.KV) []int32 {
	// 先删除原来已有的属性
	propKeys := p.RmPropKvs(key)
	pm := p.PropMap
	var ps *PropSum
	var ok bool
	keys := make([]int32, 0, len(list))
	for _, kv := range list {
		// 判断合法性
		if !checkKey(kv.Key) {
			continue
		}
		ps, ok = pm[kv.Key]
		if !ok {
			ps.Total = 0
			ps.Props = make(map[PropKey]int32)
			pm[kv.Key] = ps
		}
		ps.Total += kv.Value
		ps.Props[key] = kv.Value
		keys = append(keys, kv.Key)
	}
	p.Cache[key] = keys
	return MergerKeys(propKeys, keys)
}

func (p *PropData) RmPropKvs(key PropKey) []int32 {
	if cache, ok := p.Cache[key]; ok {
		delete(p.Cache, key)
		pm := p.PropMap
		for _, pKey := range cache {
			rmPropKey(pm, pKey, key)
		}
		return cache
	}
	return nil
}

// 通常是计算player进程的属性
func (p *PropData) CalcProps(keys []int32) []*global.PKV {
	keysMap := GenCalcKeys(keys)
	rs := make([]*global.PKV, 0, len(keysMap))
	for key := range keysMap {
		rateKey := HasRate(key)
		switch rateKey {
		case 0:
			rs = append(rs, &global.PKV{Key: key, Value: p.PropValue(key)})
		default:
			v := p.PropValue(key)
			rate := p.PropValue(rateKey)
			value := v + gutil.Ceil(float32(v)*float32((rate)/10000))
			rs = append(rs, &global.PKV{Key: key, Value: value})
		}
	}
	return rs
}

// 通常是计算map进程的属性
func (p *PropData) CalcMapProps(keys []int32, baseProp *global.PProp) []global.PKV {
	keysMap := GenCalcKeys(keys)
	rs := make([]global.PKV, 0, len(keysMap))
	for key := range keysMap {
		baseValue := GetPropValue(baseProp, key)
		rateKey := HasRate(key)
		switch rateKey {
		case 0:
			rs = append(rs, global.PKV{Key: key, Value: p.PropValue(key) + baseValue})
		default:
			v := p.PropValue(key) + baseValue
			rate := p.PropValue(rateKey)
			value := v + gutil.Ceil(float32(v)*float32((rate)/10000))
			rs = append(rs, global.PKV{Key: key, Value: value})
		}
	}
	return rs
}

func (p *PropData) PropValue(key int32) int32 {
	return p.PropMap[key].Total
}

func MergerKeys(keys, acc []int32) []int32 {
	for _, v := range keys {
		if !gutil.SliceInt32Member(v, acc) {
			acc = append(acc, v)
		}
	}
	return acc
}

func MergerKeysVal(keys []config.KV, acc []int32) []int32 {
	for _, v := range keys {
		if !gutil.SliceInt32Member(v.Key, acc) {
			acc = append(acc, v.Key)
		}
	}
	return acc
}

func MergerProps(props, acc []config.KV) []config.KV {
	for _, v := range props {
		if i, ok := findIndex(v.Key, acc); ok {
			acc[i].Value += v.Value
		} else {
			acc = append(acc, v)
		}
	}
	return acc
}

func rmPropKey(pm map[int32]*PropSum, pKey int32, key PropKey) {
	if ps, ok := pm[pKey]; ok {
		if v, ok := ps.Props[key]; ok {
			ps.Total -= v
			delete(ps.Props, key)
			pm[pKey] = ps
		}
	}
}

func checkKey(key int32) bool {
	if key >= 1 && key <= propSize {
		return true
	}
	_, ok := sp_prop_map[key]
	return ok
}

func findIndex(key int32, arr []config.KV) (int, bool) {
	for i, v := range arr {
		if v.Key == key {
			return i, true
		}
	}
	return 0, false
}

// 通过指针位置，返回属性
func GetPropValue(prop *global.PProp, key int32) int32 {
	ptr := unsafe.Pointer(prop)
	return *(*int32)(unsafe.Pointer(uintptr(ptr) + uintptr(key-1)*4))
}

func SetPropValue(prop *global.PProp, key int32, value int32) {
	ptr := unsafe.Pointer(prop)
	*(*int32)(unsafe.Pointer(uintptr(ptr) + uintptr(key-1)*4)) = value
}

func GenCalcKeys(keys []int32) map[int32]bool {
	rs := make(map[int32]bool, len(keys))
	for _, key := range keys {
		if key == PROP_MoveSpeedRate {
			rs[PROP_MoveSpeed] = true
		} else {
			rs[key] = true
			if pKey := IsRate(key); pKey > 0 {
				rs[pKey] = true
			}
		}
	}
	return rs
}

func HasRate(key int32) int32 {
	switch key {
	case PROP_MaxHP:
		return PROP_MaxHpRate
	case PROP_PhyAttack:
		return PROP_AttackRate
	case PROP_ArmorBreak:
		return PROP_ArmorBreakRate
	case PROP_PhyDefence:
		return PROP_DefenceRate
	case PROP_Hit:
		return PROP_HitAddRate
	case PROP_Miss:
		return PROP_MissAddRate
	case PROP_Crit:
		return PROP_CritAddRate
	case PROP_Tenacity:
		return PROP_TenacityAddRate
	case PROP_MoveSpeed:
		return PROP_MoveSpeedRate
	}
	return 0
}

func IsRate(key int32) int32 {
	switch key {
	case PROP_MaxHpRate:
		return PROP_MaxHP
	case PROP_AttackRate:
		return PROP_PhyAttack
	case PROP_ArmorBreakRate:
		return PROP_ArmorBreak
	case PROP_DefenceRate:
		return PROP_PhyDefence
	case PROP_HitAddRate:
		return PROP_Hit
	case PROP_MissAddRate:
		return PROP_Miss
	case PROP_CritAddRate:
		return PROP_Crit
	case PROP_TenacityAddRate:
		return PROP_Tenacity
	case PROP_MoveSpeedRate:
		return PROP_MoveSpeed
	}
	return 0
}
