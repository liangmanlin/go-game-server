package maps

import (
	"game/config"
	"game/global"
	"game/lib"
	"github.com/liangmanlin/gootp/astar"
	"github.com/liangmanlin/gootp/kernel"
	"github.com/liangmanlin/gootp/rand"
	"github.com/liangmanlin/gootp/timer"
	"unsafe"
)

const (
	AREA_WIDTH  int32 = 15
	AREA_HEIGHT int32 = 15
)

const (
	ACTOR_STATE_NORMAL int8 = iota + 1
	ACTOR_STATE_DEAD
)

const (
	MAP_MONSTER_HP int32 = iota + 6
	MAP_MONSTER_MAX_HP
	MAP_MONSTER_MOVE_SPEED
)

const (
	RETURN_TYPE_TRUE ReturnType = iota + 1
	RETURN_TYPE_FALSE
	RETURN_TYPE_TIME
)

type app struct {
	mapConfigPath string
}

type MapConfig struct {
	MapID    int32
	Width    int32
	Height   int32
	BornPos  Area
	PosList  []MapPos
	PosCount int32
	Monsters []MapConfigMonster
}

// 强制内存对齐
type MapPos struct {
	X    int16
	Y    int16
	Type int16
	G    int16
}

type MapState struct {
	*kernel.Context
	Config      *MapConfig
	Mod         *MapMod
	Name        string
	AreaMap     [global.C_ACTOR_SIZE]AreaMap
	MapRoles    map[int64]*global.PMapRole
	MapMonsters map[int64]*global.PMapMonster
	Roles       map[int64]*MapRoleData
	Monsters    map[int64]*MState
	Rand        rand.Rand
	Timer       *timer.Timer
	LC          int
	Actors      map[AKey]*MapActor
	MoveActor   map[AKey]*Move
	AStar       *astar.AStar
	DataMonster *DataMonster
	DataSkill   *DataSkill
}

type MapInfo interface {
	ID() int64
	Type() int8
	GetMove() (int32, *global.PMovePath, *global.PPos)
	SetMovePath(path *global.PMovePath)
	GetPos() *global.PPos
	IsAlive() bool
	GetRadius() float64
	GetCamp() int8
	GetBuffs() []*global.PBuff
	SetBuffs([]*global.PBuff)
}

type DataMonster struct {
	MonsterID     int64
	StopMonster   int64
	DelMonster    []int64
	PosMonsterNum map[Area]int32
	AroundDest    map[AKey][]int64
	IsMonsterLoop bool
}

type DataSkill struct {
	Index      int32
	EntityList map[int32]*SkillEntity
	Del        []int32
	Add        map[int32]*SkillEntity
	IsLoop     bool
}

type Area struct {
	X, Y int16
}

type Grid struct {
	X, Y int32
}

type MapRoleData struct {
	TcpPid   *kernel.Pid
	Player   *kernel.Pid
	MoveFail int32
	Skills   map[int32]*FightSkill
}

type FightSkill struct {
	CoolTime int64
}

type MState struct {
	ID             int64
	TypeID         int32
	State          int8
	NextTime       int64
	LastAttackTime int64
	BornPos        *global.PPos
	DeadCB         *func(state *MapState, monsterID int64, srcActor *MapActor, args interface{})
	DeadArgs       interface{}
	Doings         []*Doing
	Enemies        map[AKey]*MonsterEnemy
	Return         *MReturn
	Config         *config.DefMonster // 这里持有了一个配置，这样做对gc不友好，因为假如热更了配置，这里要等怪物死亡才会更新
	Skills         []MonsterSkill
	HPChange	   *func(state *MapState, mState *MState, srcActor *MapActor,reduce int32)
}

type MonsterSkill struct {
	SkillID int32
	CD      int32
	UseTime int64
}

type MReturn struct {
	Type ReturnType
	Time int64
}

type Doing struct {
	Fun  *behaviorFunc
	Args interface{}
	Type int32
}

type MonsterEnemy struct {
	ActorID        int64
	ActorType      int8
	TotalHurt      int64
	LastAttackTime int64
}

type MapMod struct {
	// 初始化场景回调
	Init func(state *MapState, ctx *kernel.Context, mapID int32, args ...interface{})
	// 角色进入地图时回调
	RoleEnter func(state *MapState, ctx *kernel.Context, roleID int64)
	// 角色重连回调
	RoleReconnect func(state *MapState, ctx *kernel.Context, roleID int64)
	// 角色离开地图前回调
	RoleLeave func(state *MapState, ctx *kernel.Context, roleID int64, isExit bool)
	// 判断角色是否可以进入该场景，可以为空
	RoleCanEnter func(state *MapState, ctx *kernel.Context, roleID int64) bool
	// 角色死亡回调
	RoleDead func(state *MapState, ctx *kernel.Context, roleID int64)
	// 角色复活回调
	RoleRelive func(state *MapState, ctx *kernel.Context, roleID int64)
	// 场景消息 handler
	Handle func(state *MapState, ctx *kernel.Context, msg interface{})
}

// actorKey
type AKey struct {
	ActorID   int64
	ActorType int8
}

type Move struct {
	LastUpdateTime int64 // 毫秒
	StepCount      float64
	StepTotal      float64
	DestX          float32
	DestY          float32
}

type MapActor struct {
	ActorID   int64
	ActorType int8
	State     int8
	IsMove    bool
	BaseProp  *global.PProp
	TotalProp *global.PProp
	Skills    []int32 // 考虑到同时释放的技能不会很多，这里使用slice保存
	BuffData  *BuffData
	PropData  *lib.PropData
}

type BuffData struct {
	Type2ID     map[int16]int16
	EffectCount []int16
	Data        map[int16]*BuffInfo
}

type BuffInfo struct {
	AddTime  int64
	TickTime int64
}

type MonsterCreateArg struct {
	TypeID   int32
	X, Y     float32
	Dir      int16
	Level    int16
	DeadCB   *func(state *MapState, monsterID int64, srcActor *MapActor, args interface{})
	DeadArgs interface{}
	Prop     *global.PProp
}

type MoveArg struct {
	DX, DY   float32
	IdleTime int32
}

type MoveDestArg struct {
	ActorID   int64
	ActorType int8
	Dist      int32
	DX, DY    float32
}

type MapConfigMonster struct {
	X, Y   int16
	TypeID int32
}

type ReturnType int8

type AreaFoldFunc = func(state *MapState, srcInfo, destInfo MapInfo, args ...unsafe.Pointer) bool

type SkillEntity struct {
	ID          int32 // 肯定不可能有这么多，本来是一个循环，所以溢出了也没问题
	MovePhase   int32
	EffectPhase int32
	MoveTime    int64
	EffectTime  int64
	Cfg         *config.DefSkill // 暂时先缓存配置
	MapInfo     MapInfo
	SkillPos    *global.PPos
	FlyInfo     *SkillFlyInfo  // 飞行子弹特有
	InitPos     *global.PPos   // 技能起始位置
	Target      *global.PActor // 锁定目标技能特有
}

type SkillFlyInfo struct {
	EffectActors map[AKey]bool // 保存已经造成伤害的目标，规避重复
	Frames       int32         // 记录已经飞行了多少针
	RoleCount    int32
	MonsterCount int32
}

type ChangeMapArg struct {
	RoleID      int64
	DestMapID   int32
	DestMapName string
	DestNode    *kernel.Node
	IsExit      bool
}

type MapChangeData struct {
	RoleID       int64
	MapInfo      *global.PMapRole
	Actor        *MapActor
	RoleData     *MapRoleData
	IsFirstEnter bool
}

type AgentChangeMap struct {
	MapName string
	Change  *MapChangeData
}
