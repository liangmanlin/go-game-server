package config

import (
	"encoding/json"
	"sync/atomic"
	"unsafe"
)

// 自动生成，请勿随便修改
func init() {
	//init-ptr-start
	PtrMap["Maps"] = Maps
	PtrMap["Buffs"] = Buffs
	PtrMap["SkillEffect"] = SkillEffect
	PtrMap["Skill"] = Skill
	PtrMap["MonsterBehavior"] = MonsterBehavior
	PtrMap["MonsterProp"] = MonsterProp
	PtrMap["Monster"] = Monster
	PtrMap["Goods"] = Goods
	PtrMap["BossWorld"] = BossWorld
	PtrMap["Server"] = Server
}

type KV struct {
	Key   int32
	Value int32
}

type SkillMove struct {
	StartTime int32
	Dist      int32
	EndTime   int32
}

type SkillShape struct {
	ShapeType int32
	A         float32
	B         float32
	C         float32
	D         float32
}

//BossWorld-start
type DefBossWorld struct {
	Type_id int32 // bossID
	Map_id int32 // 地图id
	Dec_lv int32 // 等级差限制
	Refresh_time int32 // 刷新时间
	X int32 // x
	Y int32 // y
	Tire int32 // 消耗体力
	Sp_drop []int32 // 第一层的boss特殊掉落
}
type bossWorldConfig struct{
	m map[int32]int
	arr []DefBossWorld
}
var BossWorld = &bossWorldConfig{}
func (B *bossWorldConfig)Get(key int32)*DefBossWorld{
	return &B.arr[B.m[key]]
}
func (B *bossWorldConfig)load(path string)  {
	c := &bossWorldConfig{m:make(map[int32]int)}
	if err:= json.Unmarshal(readFile("BossWorld",path), &c.arr);err != nil{panic(err)}
	for i:=0;i<len(c.arr);i++{
		c.m[c.arr[i].Type_id] = i
	}
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&BossWorld)),*(*unsafe.Pointer)(unsafe.Pointer(&c)))
}
func (B *bossWorldConfig)All()[]DefBossWorld {
	return B.arr
}
//BossWorld-end
//Goods-start
type DefGoods struct {
	Type_id int32 // type_id
	Name string // 名称
	Color int32 // 颜色
	Pile_num int32 // 堆叠数量
}
type goodsConfig struct{
	m map[int32]int
	arr []DefGoods
}
var Goods = &goodsConfig{}
func (G *goodsConfig)Get(key int32)*DefGoods{
	return &G.arr[G.m[key]]
}
func (G *goodsConfig)load(path string)  {
	c := &goodsConfig{m:make(map[int32]int)}
	if err:= json.Unmarshal(readFile("Goods",path), &c.arr);err != nil{panic(err)}
	for i:=0;i<len(c.arr);i++{
		c.m[c.arr[i].Type_id] = i
	}
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&Goods)),*(*unsafe.Pointer)(unsafe.Pointer(&c)))
}
func (G *goodsConfig)All()[]DefGoods {
	return G.arr
}
//Goods-end
//Monster-start
type DefMonster struct {
	Id int32 // 地图ID+XXX
	Name string // 怪物名字
	Type int32 // 类型1小怪，2个人boss，3世界boss，4通用boss，6召唤boss
	Level int32 // 等级
	Behavior int32 // 行为ID：对应monster_behavior中的ID。主动怪是1XXXX，被动怪是2XXXX
	NotReduceHP int32 // 0：扣血 1：不扣血 2：载具类型，扣血，但是不会增加仇恨
	AttackSpeed int32 // 攻击间隔
	PropID int32 // 属性id
	GuardRadius int32 // 警戒半径
	AttackRadius int32 // 攻击半径
	ActivityRadius int32 // 活动半径
	Skills []int32 // 技能（技能ID列表）：如果是延迟类的，配置为{ID,延迟时间}， 时间为毫秒
	RefreshTime int32 // 刷新时间：0不再刷新，单位是毫秒
	Drops []int32 // 掉落方案
	Exp int32 // 经验基数
	Gold int32 // 金币基数：不加就配0
	Radius int32 // 模型半径(单位是一个格子，只能整数小怪都配0）
	EventList []int32 // 事件ID列表
	FightPower int32 // 战力
}
type monsterConfig struct{
	m map[int32]int
	arr []DefMonster
}
var Monster = &monsterConfig{}
func (M *monsterConfig)Get(key int32)*DefMonster{
	return &M.arr[M.m[key]]
}
func (M *monsterConfig)load(path string)  {
	c := &monsterConfig{m:make(map[int32]int)}
	if err:= json.Unmarshal(readFile("Monster",path), &c.arr);err != nil{panic(err)}
	for i:=0;i<len(c.arr);i++{
		c.m[c.arr[i].Id] = i
	}
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&Monster)),*(*unsafe.Pointer)(unsafe.Pointer(&c)))
}
func (M *monsterConfig)All()[]DefMonster {
	return M.arr
}
//Monster-end
//MonsterProp-start
type DefMonsterProp struct {
	ID int32 // hero_type
	MaxHP int32 // 最大生命
	PhyAttack int32 // 攻击
	ArmorBreak int32 // 破甲
	PhyDefence int32 // 防御
	Hit int32 // 命中
	Miss int32 // 闪避
	Crit int32 // 暴击
	Tenacity int32 // 韧性
	MoveSpeed int32 // 移动速度
	HpRecover int32 // 生命回复
	MaxHpRate int32 // 增加万分值血量
	AttackRate int32 // 增加万分值攻击力
	ArmorBreakRate int32 // 破甲万分比
	DefenceRate int32 // 增加万分值防御值
	HitAddRate int32 // 命中万分比
	MissAddRate int32 // 闪避万分比
	CritAddRate int32 // 暴击万分比
	TenacityAddRate int32 // 韧性万分比
	DamageDeepen int32 // 伤害加深
	DamageDef int32 // 伤害减免
	HitRate int32 // 命中率
	MissRate int32 // 闪避几率
	CritRate int32 // 暴击几率
	CritDef int32 // 暴击抵抗
	CritValue int32 // 暴伤加成
	CritValueDef int32 // 暴伤减免
	ParryRate int32 // 格挡几率
	ParryOver int32 // 格挡穿透
	HuixinRate int32 // 会心几率
	HuixinDef int32 // 会心抵抗
	PvpDamageDeepen int32 // pvp伤害增加万分比
	PvpDamageDef int32 // pvp伤害减免万分比
	PveDamageDeepen int32 // pve伤害增加万分比
	PveDamageDef int32 // pve伤害减免万分比
	TotalDef int32 // 总体免伤万分比
}
type monsterPropConfig struct{
	m map[int32]int
	arr []DefMonsterProp
}
var MonsterProp = &monsterPropConfig{}
func (M *monsterPropConfig)Get(key int32)*DefMonsterProp{
	return &M.arr[M.m[key]]
}
func (M *monsterPropConfig)load(path string)  {
	c := &monsterPropConfig{m:make(map[int32]int)}
	if err:= json.Unmarshal(readFile("MonsterProp",path), &c.arr);err != nil{panic(err)}
	for i:=0;i<len(c.arr);i++{
		c.m[c.arr[i].ID] = i
	}
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&MonsterProp)),*(*unsafe.Pointer)(unsafe.Pointer(&c)))
}
func (M *monsterPropConfig)All()[]DefMonsterProp {
	return M.arr
}
//MonsterProp-end
//MonsterBehavior-start
type DefMonsterBehavior struct {
	Id int32 // 标识， 主动怪是1XXXX，被动怪是2XXXX
	Type int32 // 1:标志为需要进入战斗的行为 2：条件节点，判断通过后，会执行child 3：直接插入到doing中 4: 并列类型，会顺序执行逻辑，不能嵌套，一般只能出现在根节点
	Func string // 执行的函数实体
	Arg []int32 // 参数，具体咨询技术
	Child []int32 // 孩子节点，只有condition类型才会执行
}
type monsterBehaviorConfig struct{
	m map[int32]int
	arr []DefMonsterBehavior
}
var MonsterBehavior = &monsterBehaviorConfig{}
func (M *monsterBehaviorConfig)Get(key int32)*DefMonsterBehavior{
	return &M.arr[M.m[key]]
}
func (M *monsterBehaviorConfig)load(path string)  {
	c := &monsterBehaviorConfig{m:make(map[int32]int)}
	if err:= json.Unmarshal(readFile("MonsterBehavior",path), &c.arr);err != nil{panic(err)}
	for i:=0;i<len(c.arr);i++{
		c.m[c.arr[i].Id] = i
	}
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&MonsterBehavior)),*(*unsafe.Pointer)(unsafe.Pointer(&c)))
}
func (M *monsterBehaviorConfig)All()[]DefMonsterBehavior {
	return M.arr
}
//MonsterBehavior-end
//Skill-start
type DefSkill struct {
	Id int32 // 技能id
	SkillIndex int32 // 第几个技能
	PosType int32 // 目标点选择类型 1：根据施法时，传入的朝向计算 2：施法者自由选择(技能轮盘） 3: 锁定目标，并且伤害只有目标一个 4：施法时，以目标为施法点，伤害根据形状计算，可以有多段 5：仅仅用来触发release_cb，不会产生实体
	Dist int32 // 施法距离 通常只有pos_type=2，3，4才会判断这个字段
	SkillType int32 // 1:突进普攻 2：普攻 3：技能
	BreakType int32 // 打断类型 0：不可被打断 1：直接打断 2：如果是带位移的，会把位移执行完，打断伤害
	Monster int32 // 伤害怪物数量 0：不限制
	Role int32 // 伤害角色数量 0：不限制
	NASectionIndex int32 // 普攻段数序号
	SkillCD int32 // 技能cd（毫秒）
	MoveChangePos int32 // 位移后，是否改变施法的坐标
	SkillMove []*SkillMove // 技能位移{StartTime,Dist,LastTime} Dist:实际距离*100， StartTime，LastTime：毫秒
	SkillPhase []int32 // 伤害间隔（单位毫秒） 填的是伤害间隔，不是时间轴 如果是飞行，需要都配一个结束时间,skillFlyTime
	SkillFlyFrames []int32 // 子弹飞行帧数
	SkillShape []*SkillShape // 技能形状的初始参数{ShapeType,,A,B,C,D} ShapeType,1.矩形,2.圆形,3.扇形 A,B,C 参数根据不同形状不同  1.矩形{1,偏移，长，宽}（飞行子弹特殊说明：{1,偏移，长，宽，每0.1秒宽变化系数}，变化系数=0表示不变，所有子弹默认没有偏移值） 2.圆形{2，半径,中心偏移量}（没有偏移量的默认是0） 3.扇形{3，半径，圆心角}
	SkillSpread []int32 // 技能展开方式 1.中心延射线发出，中心随时间偏移 （理论上仅仅支持矩形）  2.技能区域无展开,仅维持初始区域  
	SkillSpreadSpeed []*KV // 技能展开速度(正数与技能移动方向相同，负数为相反)（每一毫秒的速度（*1000），10表示0.1秒移动1个格子 当飞行子弹时配置，{速度，类型：0：表示每段都从技能初始位置开始，1：表示每段技能都从当前飞行位置开始}
	ContinueTime int32 // 技能持续时间（毫秒）,通常是给ai用的
	Effect [][]int32 // 技能效果
	AfterEffect []int32 // 释放后叠加效果效果
	PhaseCB [][]int32 // 伤害段数回调
	FlyEndCB []int32 // 飞行结束回调
	ReleaseEffect []int32 // 施法技能时触发 是一个列表[]
	AIRange float32 // 施法时自动靠近目标的格子距离（仅对玩家有效）
}
type skillConfig struct{
	m map[int32]int
	arr []DefSkill
}
var Skill = &skillConfig{}
func (S *skillConfig)Get(key int32)*DefSkill{
	return &S.arr[S.m[key]]
}
func (S *skillConfig)load(path string)  {
	c := &skillConfig{m:make(map[int32]int)}
	if err:= json.Unmarshal(readFile("Skill",path), &c.arr);err != nil{panic(err)}
	for i:=0;i<len(c.arr);i++{
		c.m[c.arr[i].Id] = i
	}
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&Skill)),*(*unsafe.Pointer)(unsafe.Pointer(&c)))
}
func (S *skillConfig)All()[]DefSkill {
	return S.arr
}
//Skill-end
//SkillEffect-start
type DefSkillEffect struct {
	Id int32 // id
	Type int32 // 分类，1：回调函数，2：伤害函数
	Func string // 执行的函数实体
	Arg []int32 // 参数，具体咨询技术
}
type skillEffectConfig struct{
	m map[int32]int
	arr []DefSkillEffect
}
var SkillEffect = &skillEffectConfig{}
func (S *skillEffectConfig)Get(key int32)*DefSkillEffect{
	return &S.arr[S.m[key]]
}
func (S *skillEffectConfig)load(path string)  {
	c := &skillEffectConfig{m:make(map[int32]int)}
	if err:= json.Unmarshal(readFile("SkillEffect",path), &c.arr);err != nil{panic(err)}
	for i:=0;i<len(c.arr);i++{
		c.m[c.arr[i].Id] = i
	}
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&SkillEffect)),*(*unsafe.Pointer)(unsafe.Pointer(&c)))
}
func (S *skillEffectConfig)All()[]DefSkillEffect {
	return S.arr
}
//SkillEffect-end
//Buffs-start
type DefBuffs struct {
	Id int32 // id（要小于30001）
	BuffType int32 // buff_type,前100系统保留做特殊用途 要小于30001
	BuffLV int32 // 等级
	BuffValue int32 // 数值，某些buff这个配置没意义，例如无敌
	LastTime int32 // 持续时间（毫秒）,负数表示永久buff
	EffectType int32 // 效果类型，区别于buff_type，这里是统一效果，例如1：加属性
	EffectFun string // 效果函数，具体问技术
	SumType int32 // 累加类型，1：数值累加，2：时间累加,3:替换，需要保证buff没有特殊数据逻辑,4,替换有变化的（属性需要配成4），5，晕眩递减
	InvType int32 // 间隔类型，1：普通，2，间隔作用
	Inv int32 // 间隔时间（毫秒）
	OpFun string // buff添加，删除时回调
	IsFight int32 // 是否战斗buff，战斗buff下线会清除
	IsDebuff int32 // 是否负面效果（可以被驱散）
	NoticeType int32 // 1:广播通知 2：只通知自己 0：不通知
	ChangeMapDel int32 // 跳转场景是否删除
	AvoidList []int32 // 会被哪些buff_type免疫
	CleanList []int32 // 添加buff时，清除buff_type 通常不需要配
	NoMmove int32 // 有这个buff时，不能移动
	NoSkill int32 // 有这个buff时，不能释放技能，并且打断施法
}
type buffsConfig struct{
	m map[int32]int
	arr []DefBuffs
}
var Buffs = &buffsConfig{}
func (B *buffsConfig)Get(key int32)*DefBuffs{
	return &B.arr[B.m[key]]
}
func (B *buffsConfig)load(path string)  {
	c := &buffsConfig{m:make(map[int32]int)}
	if err:= json.Unmarshal(readFile("Buffs",path), &c.arr);err != nil{panic(err)}
	for i:=0;i<len(c.arr);i++{
		c.m[c.arr[i].Id] = i
	}
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&Buffs)),*(*unsafe.Pointer)(unsafe.Pointer(&c)))
}
func (B *buffsConfig)All()[]DefBuffs {
	return B.arr
}
//Buffs-end
//Maps-start
type DefMaps struct {
	Id int32 // 地图id,不能大于3999（规则：地图ID一般为三位数。主线野外地图为1开头，主线副本为2开头，Boss地图为3开头并以十位区分类型，个位区分层数，玩法地图为4开头，功能地图如帮派领地、结婚场景5开头、秘境地图6开头）
	Type int32 // 类型
	Level int32 // 等级
	BlockPath int32 // 阻挡文件编号
	PkModel []int32 // 允许的pk模式  1）和平 2）强制 3）全体 4）同服 5）仇人 6）阵营
	BornX int32 // 出生点
	BornY int32 // 出生点
}
type mapsConfig struct{
	m map[int32]int
	arr []DefMaps
}
var Maps = &mapsConfig{}
func (M *mapsConfig)Get(key int32)*DefMaps{
	return &M.arr[M.m[key]]
}
func (M *mapsConfig)load(path string)  {
	c := &mapsConfig{m:make(map[int32]int)}
	if err:= json.Unmarshal(readFile("Maps",path), &c.arr);err != nil{panic(err)}
	for i:=0;i<len(c.arr);i++{
		c.m[c.arr[i].Id] = i
	}
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&Maps)),*(*unsafe.Pointer)(unsafe.Pointer(&c)))
}
func (M *mapsConfig)All()[]DefMaps {
	return M.arr
}
//Maps-end
