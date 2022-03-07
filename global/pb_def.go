package global

//--------------------------Login------------------------------------------------------------

// 连接
type LoginTosConnect struct {
	IsReconnect bool
	Token       string
	RecID       int64
}
type LoginTocConnect struct {
	Succ        bool
	IsReconnect bool
	Token       string
}

// 登录协议
type LoginTosLogin struct {
	Account  string
	AgentID  int32
	ServerID int32
}
type LoginTocLogin struct {
	Succ       bool
	Reason     *PMsg
	Account    string
	LastRoleID int64
	Roles      []*PRole
}

type LoginTosCreateRole struct {
	Name     string
	HeroType int32
	Sex      int32
}
type LoginTocCreateRole struct {
	Succ   bool
	Reason *PMsg
	RoleID int64
}

type LoginTosSelect struct {
	RoleID int64
}
type LoginTocSelect struct {
	Succ   bool
	Reason *PMsg
	RoleID int64
}

//--------------------------Game------------------------------------------------------------

type GameTosUP struct { // router:LoginLogin

}

type GameTocUP struct {
}

//-----------------------map---------------------------------------------------------------
type MapTocActorLeaveArea struct {
	ActorID   int64
	ActorType int8
}

type MapTocEnterArea struct {
	RolesShow    []*PMapRole
	MonstersShow []*PMapMonster
	RolesDel     []int64
	MonstersDel  []int64
}

type MapTocRoleEnterArea struct {
	MapInfo *PMapRole
}

type MapTocMonsterEnterArea struct {
	MapInfo *PMapMonster
}

type MapTocStop struct {
	ActorID   int64
	ActorType int8
	Dir       int16
	X         float32
	Y         float32
}

type MapTosMove struct {
}
type MapTocMove struct {
	ActorID   int64
	ActorType int8
	StartX    float32
	StartY    float32
	MovePath  *PMovePath
}

type MapTosMoveStop struct {
	X   float32
	Y   float32
	Dir int16
}
type MapTocMoveStop struct {
	ActorID   int64
	ActorType int8
	Dir       int16
	X         float32
	Y         float32
}

type MapTocUpdateMonsterInfo struct {
	ID int64
	UL []*PKV
}

type MapTocUpdateRoleInfo struct {
	RoleID int64
	UL     []*PKV
}

type MapTocActorDead struct {
	ActorID   int64
	ActorType int8
}

type MapTocDelBuff struct {
	ActorID   int64
	ActorType int8
	BuffID    int16
}

//------------------------------------map end--------------------------------------------
//------------------------------------fight----------------------------------------------

type FightTocUseSkill struct {
	ActorID   int64
	ActorType int8
	SkillID   int32
	X         float32
	Y         float32
}

type FightTocSkillEffect struct {
	SkillID    int32
	SkillDir   int16
	SrcType    int8
	SrcID      int64
	EffectList []*PSkillEffect
}

//------------------------------------fight end------------------------------------------

//-------------------------------------role----------------------------------------------
type RoleTocUPProps struct {
	UP []*PKV
}

//-------------------------------------role end-------------------------------------------

type PMsg struct {
	MsgID int32
	Bin   []byte // 客户端需要根据商定的规则解开一个字符参数
}

type PRole struct {
	RoleID   int64
	RoleName string
	HeroType int32
	Level    int32
	Skin     *PSkin
}

type PSkin struct {
}

type PGoods struct {
	RoleID     int64
	ID         int32
	Type       int32
	TypeID     int32
	Num        int32
	Bind       bool
	StartTime  int32
	EndTime    int32
	CreateTime int32
}

type PRoleAttr struct {
	RoleID  int64
	Diamond int64
	Gold    int64
}


type PRoleBase struct {
	RoleID   int64
	Name     string
	Account  string
	AgentID  int32
	ServerID int32
	HeroType int32
	Level    int32
	Skin     *PSkin
}

type PPos struct {
	X   float32
	Y   float32
	Dir int16
}

type PMapRole struct {
	RoleID    int64
	Name      string
	HeroType  int32
	ServerID  int32
	Level     int16
	State     int8
	Camp      int8
	Skin      *PSkin
	Pos       *PPos
	MovePath  *PMovePath
	HP        int32
	MaxHP     int32
	MoveSpeed int32
	Buffs     []*PBuff
}

type PMapMonster struct {
	MonsterID int64
	TypeID    int32
	Level     int16
	Radius    int8
	Camp      int8
	Pos       *PPos
	MovePath  *PMovePath
	HP        int32
	MaxHP     int32
	MoveSpeed int32
	Buffs     []*PBuff
}

type PBuff struct {
	ID      int16
	Type    int16
	SrcType int8
	SrcID   int64
	EndTime int64
	Value   int32
}

type PMovePath struct {
	EndX     float32
	EndY     float32
	GridPath []*PGrid
}

type PGrid struct {
	GridX int16
	GridY int16
}

type PProp struct {
	MaxHP           int32
	PhyAttack       int32
	ArmorBreak      int32
	PhyDefence      int32
	Hit             int32
	Miss            int32
	Crit            int32
	Tenacity        int32
	MoveSpeed       int32
	HpRecover       int32
	MaxHpRate       int32
	AttackRate      int32
	ArmorBreakRate  int32
	DefenceRate     int32
	HitAddRate      int32
	MissAddRate     int32
	CritAddRate     int32
	TenacityAddRate int32
	DamageDeepen    int32
	DamageDef       int32
	HitRate         int32
	MissRate        int32
	CritRate        int32
	CritDef         int32
	CritValue       int32
	CritValueDef    int32
	ParryRate       int32
	ParryOver       int32
	HuixinRate      int32
	HuixinDef       int32
	PvpDamageDeepen int32
	PvpDamageDef    int32
	PveDamageDeepen int32
	PveDamageDef    int32
	TotalDef        int32
}

type PKV struct {
	Key   int32
	Value int32
}

type PActor struct {
	ActorID   int64
	ActorType int8
}

type PSkillEffect struct {
	ActorID    int64
	ActorType  int8
	EffectType int8
	Value      int32
}

type PCeilFP struct {
	ID         int32
	SubID      int32
	FightPower int64
}
