package lib

type PropData struct {
	Cache   map[PropKey][]int32
	PropMap map[int32]*PropSum
}

type PropKey struct {
	ID    int32
	SubID int32
}

type PropSum struct {
	Total int32
	Props map[PropKey]int32
}

const (
	PROP_MaxHP           int32 = iota+1
	PROP_PhyAttack
	PROP_ArmorBreak
	PROP_PhyDefence
	PROP_Hit
	PROP_Miss
	PROP_Crit
	PROP_Tenacity
	PROP_MoveSpeed
	PROP_HpRecover
	PROP_MaxHpRate
	PROP_AttackRate
	PROP_ArmorBreakRate
	PROP_DefenceRate
	PROP_HitAddRate
	PROP_MissAddRate
	PROP_CritAddRate
	PROP_TenacityAddRate
	PROP_DamageDeepen
	PROP_DamageDef
	PROP_HitRate
	PROP_MissRate
	PROP_CritRate
	PROP_CritDef
	PROP_CritValue
	PROP_CritValueDef
	PROP_ParryRate
	PROP_ParryOver
	PROP_HuixinRate
	PROP_HuixinDef
	PROP_PvpDamageDeepen
	PROP_PvpDamageDef
	PROP_PveDamageDeepen
	PROP_PveDamageDef
	PROP_TotalDef
)

// 添加是需要处理下面的一个map
const (
	PROP_MoveSpeedRate int32 = iota+1000
)

var sp_prop_map = map[int32]int32{
	PROP_MoveSpeedRate:PROP_MoveSpeed,
}

const (
	PROP_KEY_BASE int32 = iota+1
	PROP_KEY_LEVEL
	PROP_KEY_BUFF
)
