package maps

import (
	"game/config"
	"game/global"
	"game/lib"
	"github.com/liangmanlin/gootp/kernel"
	"math"
	"runtime/debug"
	"sync"
	"unsafe"
)

const FRAMES_TIME int32 = 100

const (
	SKILL_POS_TYPE_1 int32 = iota + 1 // 根据施法时，传入的朝向计算,起始坐标是施法者坐标
	SKILL_POS_TYPE_2                  // 施法者自由选择(技能轮盘）
	SKILL_POS_TYPE_3                  // 锁定目标，并且伤害只有目标一个
	SKILL_POS_TYPE_4                  // 施法时，以目标为施法点，伤害根据形状计算，可以有多段
	SKILL_POS_TYPE_5                  // 仅仅用来触发release_cb，不会产生实体
)

const (
	SHAPE_TYPE_RECTANGLE int32 = iota + 1 // 矩形
	SHAPE_TYPE_CIRCLE                     // 圆
	SHAPE_TYPE_SECTOR                     // 扇形
)

const SPREAD_FLY int32 = 1 // 飞行

var _skillPosPool = sync.Pool{
	New: func() interface{} {
		return &global.PPos{}
	},
}

func NewPos() *global.PPos {
	return _skillPosPool.Get().(*global.PPos)
}

var _skillEntityPool = sync.Pool{
	New: func() interface{} {
		return &SkillEntity{}
	},
}

func MakeSkillEntity() *SkillEntity {
	return _skillEntityPool.Get().(*SkillEntity)
}

var _skillFlyPool = sync.Pool{
	New: func() interface{} {
		return &SkillFlyInfo{}
	},
}

func NewFlyInfo() *SkillFlyInfo {
	r := _skillFlyPool.Get().(*SkillFlyInfo)
	r.Frames = 0
	if r.EffectActors != nil {
		for k := range r.EffectActors {
			delete(r.EffectActors, k)
		}
	} else {
		r.EffectActors = make(map[AKey]bool, 10)
	}
	r.RoleCount = 0
	r.MonsterCount = 0
	return r
}

var _skillEffectPool = sync.Pool{
	New: func() interface{} {
		return make([]*global.PSkillEffect, 0, 10)
	},
}

func NewEffectList() []*global.PSkillEffect {
	s := _skillEffectPool.Get().([]*global.PSkillEffect)
	return s[0:0]
}

func ReleaseEffectList(l []*global.PSkillEffect) {
	for _, e := range l {
		_skillEffectOnePool.Put(e)
	}
	_skillEffectPool.Put(l)
}

var _skillEffectOnePool = sync.Pool{
	New: func() interface{} {
		return &global.PSkillEffect{}
	},
}

func NewEffect(actorType int8, actorID int64, effectType int8, damage int32) *global.PSkillEffect {
	s := _skillEffectOnePool.Get().(*global.PSkillEffect)
	s.EffectType = effectType
	s.Value = damage
	s.ActorID = actorID
	s.ActorType = actorType
	return s
}

var RoleReleaseSkill = func(state *MapState, mapInfo *global.PMapRole, skillID int32, x, y float32, dir int16, target *global.PActor) interface{} {
	if !mapInfo.IsAlive() {
		return lib.NewPMsg(101)
	}
	cfg := config.Skill.Get(skillID)
	curPos := mapInfo.Pos
	skillPos := NewPos()
	switch cfg.PosType {
	case SKILL_POS_TYPE_1:
		skillPos.X = curPos.X
		skillPos.Y = curPos.Y
	case SKILL_POS_TYPE_2:
		// 检查施法距离
		if !lib.XYLessThan(curPos.X, curPos.Y, x, y, float32(cfg.Dist)) {
			return lib.NewPMsg(102)
		}
		skillPos.X = x
		skillPos.Y = y
	case SKILL_POS_TYPE_5:
		skillPos.X = x
		skillPos.Y = y
	default:
		// 其余的应该是检查目标距离
		dPos := state.GetMapInfo(target.ActorType, target.ActorID).GetPos()
		if !lib.PosLessThan(curPos, dPos, float32(cfg.Dist)) {
			return lib.NewPMsg(102)
		}
		skillPos.X = x
		skillPos.Y = y
	}
	skillPos.Dir = dir
	now2 := kernel.Now2()
	// 判断cd
	if !CheckCoolTime(state, mapInfo.RoleID, now2, cfg) {
		return lib.NewPMsg(103)
	}
	proto := &global.FightTocUseSkill{
		ActorID:   mapInfo.RoleID,
		ActorType: global.ACTOR_ROLE,
		SkillID:   skillID,
		X:         skillPos.X,
		Y:         skillPos.Y,
	}
	// 这里没有在施法者的位置广播，会一些可能在远方的人，这里选择施法坐标比较合理
	state.BroadCastPos(skillPos, proto)
	actor := state.GetActorByID(global.ACTOR_ROLE, mapInfo.RoleID)
	skillEntity := NewSkillEntity(state, mapInfo, actor, skillPos, cfg, target, now2)
	TryActSkill(state, skillEntity, now2)
	return global.OK{}
}

var MonsterReleaseSKill = func(state *MapState, mapInfo *global.PMapMonster, skillID int32, x, y float32, dir int16,
	target *global.PActor,now2 int64) {
	if !mapInfo.IsAlive() {
		return
	}
	cfg := config.Skill.Get(skillID)
	curPos := mapInfo.Pos
	skillPos := NewPos()
	switch cfg.PosType {
	case SKILL_POS_TYPE_4:
		skillPos.X, skillPos.Y = x, y
	default:
		skillPos.X,skillPos.Y = curPos.X,curPos.Y
	}
	skillPos.Dir = dir
	proto := &global.FightTocUseSkill{
		ActorID:   mapInfo.MonsterID,
		ActorType: global.ACTOR_MONSTER,
		SkillID:   skillID,
		X:         skillPos.X,
		Y:         skillPos.Y,
	}
	state.BroadCastPos(skillPos, proto)
	actor := state.GetActorByID(global.ACTOR_MONSTER, mapInfo.MonsterID)
	skillEntity := NewSkillEntity(state, mapInfo, actor, skillPos, cfg, target, now2)
	TryActSkill(state, skillEntity, now2)
}

func NewSkillEntity(state *MapState, mapInfo MapInfo, actor *MapActor, skillPos *global.PPos,
	cfg *config.DefSkill, target *global.PActor, now2 int64) *SkillEntity {
	entity := MakeSkillEntity()
	entity.Cfg = cfg
	if cfg.PosType == SKILL_POS_TYPE_5 {
		entity.ID = 0
	} else {
		entity.ID = MakeSkillEntityID(state)
		actor.Skills = append(actor.Skills, entity.ID)
	}
	if len(cfg.SkillMove) > 0 {
		entity.MovePhase = 0
		entity.MoveTime = now2 + int64(cfg.SkillMove[0].StartTime)
	} else {
		entity.MoveTime = 0
	}
	if len(cfg.SkillPhase) > 0 {
		entity.EffectPhase = 0
		entity.EffectTime = now2 + int64(cfg.SkillPhase[0])
	} else {
		entity.EffectTime = 0
	}
	if len(cfg.SkillFlyFrames) > 0 && cfg.SkillFlyFrames[0] > 0 {
		// 飞行子弹
		entity.FlyInfo = NewFlyInfo()
	} else {
		entity.FlyInfo = nil
	}
	entity.InitPos = NewPos()
	*entity.InitPos = *skillPos
	entity.SkillPos = skillPos
	entity.Target = target
	entity.MapInfo = mapInfo
	if len(cfg.ReleaseEffect) > 0 {
		ActiveCallBack(state, entity, cfg.ReleaseEffect, nil)
	}
	return entity
}

func TryActSkill(state *MapState, entity *SkillEntity, now2 int64) {
	cfg := entity.Cfg
	switch cfg.PosType {
	case SKILL_POS_TYPE_3: // 锁定目标技能
		time := cfg.SkillPhase[0]
		if time <= 150 {
			// 短时间的技能可以立即施法
			if EntityUpdate(state, entity, now2+200) {
				state.AddSkillEntity(entity)
			}
		}
	case SKILL_POS_TYPE_5:
	default:
		state.AddSkillEntity(entity)
	}
}

// 判断技能cd，成功就设置cd，并且返回 true
var CheckCoolTime = func(state *MapState, roleID int64, now2 int64, cfg *config.DefSkill) bool {
	roleData := state.Roles[roleID]
	if skill, ok := roleData.Skills[cfg.Id]; ok {
		if now2 >= skill.CoolTime {
			skill.CoolTime = now2 + int64(cfg.SkillCD)
		} else {
			return false
		}
	} else {
		return false
	}
	return true
}

func SkillUpdate(state *MapState, now2 int64) {
	data := state.DataSkill
	data.IsLoop = true
	el := data.EntityList
	var curID int32
	var finish bool
	defer catchSkill(state, &curID, &finish, el)
	var e *SkillEntity
	for curID, e = range el {
		if !EntityUpdate(state, e, now2) {
			e.Release(state)
			delete(el, curID)
		}
	}
	finish = true
	data.IsLoop = false
	add := data.Add
	for ID, e := range add {
		el[ID] = e
		delete(add, ID)
	}
	size := len(data.Del)
	if size > 0 {
		for _, ID := range data.Del {
			if e, ok := el[ID]; ok {
				e.Release(state)
				delete(el, ID)
			}
		}
		data.Del = data.Del[0:0]
	}
}

func catchSkill(state *MapState, curID *int32, finish *bool, el map[int32]*SkillEntity) {
	if *finish {
		return
	}
	p := recover()
	if p != nil {
		e := el[*curID]
		kernel.ErrorLog("catch skill error:%s,Stack:%s\nentity:%s", p, debug.Stack(), lib.ToString(e))
		e.Release(state)
		delete(el, *curID)
	}
}

func EntityUpdate(state *MapState, entity *SkillEntity, now2 int64) bool {
	if entity.MoveTime > 0 && now2 >= entity.MoveTime {
		EntityMove(state, entity, now2)
	}
	if entity.EffectTime > 0 && now2 >= entity.EffectTime {
		// 执行效果
		EntityEffect(state, entity, now2)
	}
	if entity.EffectTime > 0 || entity.MoveTime > 0 {
		// 任意一个效果都要继续
		return true
	}
	return false
}

var EntityEffect = func(state *MapState, entity *SkillEntity, now2 int64) {
	srcMi := entity.MapInfo // 如果程序没有错误，这里取到的人一定还在这个地图上
	if !srcMi.IsAlive() {
		entity.EffectTime = 0
		return
	}
	phase := entity.EffectPhase
	cfg := entity.Cfg
	srcPos := srcMi.GetPos()
	targets := MakeMapInfoSlice()
	var dir int16
	if cfg.PosType == SKILL_POS_TYPE_3 {
		// 锁定目标，直接得出
		tg := entity.Target
		if mi := state.GetMapInfo(tg.ActorType, tg.ActorID); mi != nil {
			targets = append(targets, state.GetMapInfo(tg.ActorType, tg.ActorID))
		} else {
			entity.EffectTime = 0
			return
		}
		dir = srcPos.Dir
	} else {
		// 需要通过参数，计算命中的目标
		effectShape := cfg.SkillShape[phase]
		effectSpread := cfg.SkillSpread[phase]
		currentPos, shape, width, height := CalcCurrentParam(entity, effectShape, effectSpread, phase)
		dir = currentPos.Dir
		srcPos.Dir = dir // 需要修改角色的朝向，不需要额外同步，技能伤害协议可以同步
		// 先分表取出目标，判断pk模式之类
		areaList := Get9AreaByArea(GetAreaByXY(currentPos.X, currentPos.Y))
		monsterTargets := AreasActorFold(state, global.ACTOR_MONSTER, areaList, srcMi, SkillSearchPK, unsafe.Pointer(entity))
		rolesTargets := AreasActorFold(state, global.ACTOR_ROLE, areaList, srcMi, SkillSearchPK, unsafe.Pointer(entity))
		// 判断目标是否在范围内
		monsterTargets = CollisionMath(state, currentPos, shape, width, height, srcMi, monsterTargets, cfg.Monster)
		rolesTargets = CollisionMath(state, currentPos, shape, width, height, srcMi, rolesTargets, cfg.Role)
		targets = append(targets, monsterTargets...)
		targets = append(targets, rolesTargets...)
		// 归还池
		ReleaseMapInfoSlice(monsterTargets)
		ReleaseMapInfoSlice(rolesTargets)
	}
	srcActor := state.GetActorByID(srcMi.Type(), srcMi.ID())
	// 执行效果
	effectList := NewEffectList()
	for _, effectID := range cfg.Effect[phase] {
		effectList = CalcEffect(state, entity, srcActor, effectID, targets, effectList)
	}
	if len(effectList) > 0 {
		proto := &global.FightTocSkillEffect{
			SkillID:    cfg.Id,
			SrcType:    srcMi.Type(),
			SrcID:      srcMi.ID(),
			SkillDir:   dir,
			EffectList: effectList,
		}
		state.BroadCastMapAreas(EffectToAreas(state, targets), proto)
	}
	if len(cfg.AfterEffect) > 0{
		// 执行效果
		ActiveCallBack(state,entity,cfg.AfterEffect,targets)
	}

	if flyInfo := entity.FlyInfo; flyInfo == nil {
		if phase < int32(len(cfg.PhaseCB)){
			ActiveCallBack(state,entity,cfg.PhaseCB[phase],targets)
		}
		NextPhase(entity, cfg, phase)
	} else {
		// 飞行子弹
		count := cfg.SkillFlyFrames[phase]
		flyInfo.Frames++
		if flyInfo.Frames >= count {
			if len(cfg.FlyEndCB) > 0{
				// 执行效果
				ActiveCallBack(state,entity,cfg.FlyEndCB,targets)
			}
			// 结束
			_skillFlyPool.Put(flyInfo)
			entity.FlyInfo = nil
			NextPhase(entity, cfg, phase)
		} else {
			// 继续飞，更新一下对象
			em := flyInfo.EffectActors
			for i := range targets {
				t := targets[i]
				em[AKey{t.ID(), t.Type()}] = true
			}
		}
	}
	// 归还池
	ReleaseMapInfoSlice(targets)
	ReleaseEffectList(effectList)
}

func NextPhase(entity *SkillEntity, cfg *config.DefSkill, phase int32) {
	phase++
	if int(phase) < len(cfg.SkillPhase) {
		entity.EffectPhase = phase
		entity.EffectTime += int64(cfg.SkillPhase[phase])
		if int(phase) < len(cfg.SkillFlyFrames) && cfg.SkillFlyFrames[phase] > 0 {
			entity.FlyInfo = NewFlyInfo()
			fly := cfg.SkillSpreadSpeed[phase]
			if fly.Value == 0 {
				// 初始化为起始坐标
				*entity.SkillPos = *entity.InitPos
			}
		}
	} else {
		// 结束
		entity.EffectTime = 0
	}
}

// 这个函数比较关键，不太能热更
func CollisionMath(state *MapState, currentPos *global.PPos, shape int32, width, height float64,
	srcMi MapInfo, targets []MapInfo, limit int32) []MapInfo {
	switch shape {
	case SHAPE_TYPE_RECTANGLE:
		return RectangleMatch(state, currentPos, width, height, srcMi, targets, limit)
	case SHAPE_TYPE_CIRCLE:
		return CircleMatch(state, currentPos, width, srcMi, targets, limit)
	default:
		return SectorMatch(state, currentPos, width, width, srcMi, targets, limit)
	}
}

func RectangleMatch(state *MapState, currentPos *global.PPos, width, height float64,
	srcMi MapInfo, targets []MapInfo, limit int32) []MapInfo {
	dir := currentPos.Dir % 360
	width = width / 2
	height = height / 2
	switch dir {
	case 0, 180:
		return rectangleNormal(state, currentPos, width, height, srcMi, targets, limit)
	case 90, 270:
		return rectangleNormal(state, currentPos, height, width, srcMi, targets, limit)
	}
	// 计算旋转矩阵
	cos := lib.Cos(dir)
	sin := lib.Sin(dir)
	size := len(targets)
	var x, y, destRadius float64
	var count int32
	srcType := srcMi.Type()
	srcPos := srcMi.GetPos()
	// 暂时不进行排序，因为会比较消耗性能
	for i := 0; (count < limit || limit == 0) && i < size; i++ {
		destMi := targets[i]
		destPos := destMi.GetPos()
		destType := destMi.Type()
		// 检测一下安全区
		if srcType == global.ACTOR_ROLE && destType == global.ACTOR_ROLE {
			if state.PosSafe(srcPos) || state.PosSafe(destPos) {
				continue
			}
		}
		destRadius = destMi.GetRadius()
		x = float64(destPos.X - currentPos.X)
		y = float64(destPos.Y - currentPos.Y)
		if math.Abs(x*cos+y*sin) <= float64(width)+destRadius && math.Abs(y*cos-x*sin) <= float64(height)+destRadius {
			targets[count] = destMi
			count++
		}
	}
	return targets[0:count]
}

func rectangleNormal(state *MapState, currentPos *global.PPos, width, height float64,
	srcMi MapInfo, targets []MapInfo, limit int32) []MapInfo {
	var x, y, destRadius float64
	var count int32
	size := len(targets)
	srcType := srcMi.Type()
	srcPos := srcMi.GetPos()
	// 暂时不进行排序，因为会比较消耗性能
	for i := 0; (count < limit || limit == 0) && i < size; i++ {
		destMi := targets[i]
		destPos := destMi.GetPos()
		destType := destMi.Type()
		// 检测一下安全区
		if srcType == global.ACTOR_ROLE && destType == global.ACTOR_ROLE {
			if state.PosSafe(srcPos) || state.PosSafe(destPos) {
				continue
			}
		}
		destRadius = destMi.GetRadius()
		x = float64(destPos.X - currentPos.X)
		y = float64(destPos.Y - currentPos.Y)
		if math.Abs(x) <= width+destRadius && math.Abs(y) <= height+destRadius {
			targets[count] = destMi
			count++
		}
	}
	return targets[0:count]
}

func CircleMatch(state *MapState, currentPos *global.PPos, radius float64,
	srcMi MapInfo, targets []MapInfo, limit int32) []MapInfo {
	var x, y, destRadius float64
	var count int32
	size := len(targets)
	srcType := srcMi.Type()
	srcPos := srcMi.GetPos()
	// 暂时不进行排序，因为会比较消耗性能
	for i := 0; (count < limit || limit == 0) && i < size; i++ {
		destMi := targets[i]
		destPos := destMi.GetPos()
		destType := destMi.Type()
		// 检测一下安全区
		if srcType == global.ACTOR_ROLE && destType == global.ACTOR_ROLE {
			if state.PosSafe(srcPos) || state.PosSafe(destPos) {
				continue
			}
		}
		destRadius = destMi.GetRadius() + radius
		x = float64(destPos.X - currentPos.X)
		y = float64(destPos.Y - currentPos.Y)
		if x*x+y*y <= destRadius*destRadius {
			targets[count] = destMi
			count++
		}
	}
	return targets[0:count]
}

// 可能使用代数方法更快，而且代码会少一点
func SectorMatch(state *MapState, currentPos *global.PPos, radius, tangle float64,
	srcMi MapInfo, targets []MapInfo, limit int32) []MapInfo {
	var sx, sy, dx, dy, totalRadius, actorRadius, dist float64
	angle := int16(tangle) / 2
	dir := currentPos.Dir
	sx, sy = float64(currentPos.X), float64(currentPos.Y)
	dirLeft := (dir + angle) % 360
	dirRight := (dir - angle + 360) % 360
	// 计算两个端点坐标
	posLeftX := sx + radius*lib.Cos(dirLeft)
	posLeftY := sy + radius*lib.Sin(dirLeft)

	posRightX := sx + radius*lib.Cos(dirRight)
	posRightY := sy + radius*lib.Sin(dirRight)

	var count int32
	var absDirLeft, absDirRight int16
	size := len(targets)
	srcType := srcMi.Type()
	srcPos := srcMi.GetPos()
	// 暂时不进行排序，因为会比较消耗性能
	for i := 0; (count < limit || limit == 0) && i < size; i++ {
		destMi := targets[i]
		destPos := destMi.GetPos()
		destType := destMi.Type()
		// 检测一下安全区
		if srcType == global.ACTOR_ROLE && destType == global.ACTOR_ROLE {
			if state.PosSafe(srcPos) || state.PosSafe(destPos) {
				continue
			}
		}
		actorRadius = destMi.GetRadius()
		dx = float64(destPos.X) - sx
		dy = float64(destPos.Y) - sy
		totalRadius = radius + actorRadius
		dist = dx*dx + dy*dy
		if dist < totalRadius*totalRadius {
			// 在圆内的才可能相交
			actorRadius = actorRadius * actorRadius
			if dist <= actorRadius {
				targets[count] = destMi
				count++
				continue
			}
			tmpDir := lib.Dir(dx, dy)
			if dirLeft > dirRight && tmpDir >= dirRight && tmpDir <= dirLeft {
				targets[count] = destMi
				count++
				continue
			}
			absDirLeft = getDiffDegree(tmpDir, dirLeft)
			absDirRight = getDiffDegree(tmpDir, dirRight)
			//取一条最近的边去计算
			if absDirLeft < absDirRight && absDirLeft < 90 {
				if lib.Sin(absDirLeft)*dist < actorRadius {
					if lib.Cos(absDirLeft)*dist < radius*radius {
						targets[count] = destMi
						count++
					} else {
						//如果端点到圆心的距离小于圆半径，仍然相交
						dx = float64(destPos.X) - posLeftX
						dy = float64(destPos.Y) - posLeftY
						if dx*dx+dy*dy <= actorRadius {
							targets[count] = destMi
							count++
						}
					}
				}
			} else if absDirLeft >= absDirRight && absDirRight < 90 {
				if lib.Sin(absDirRight)*dist < actorRadius {
					if lib.Cos(absDirRight)*dist < radius*radius {
						targets[count] = destMi
						count++
					} else {
						//如果端点到圆心的距离小于圆半径，仍然相交
						dx = float64(destPos.X) - posRightX
						dy = float64(destPos.Y) - posRightY
						if dx*dx+dy*dy <= actorRadius {
							targets[count] = destMi
							count++
						}
					}
				}
			}
		}
	}
	return targets[0:count]
}

func getDiffDegree(DirA, DirB int16) int16 {
	var a, b int16
	if DirA < DirB {
		a = DirB - DirA
		b = DirA + 360 - DirB
	} else {
		a = DirA - DirB
		b = DirB + 360 - DirA
	}
	if a < b {
		return a
	}
	return b
}

// 如果施法目标需要移动
var EntityMove = func(state *MapState, entity *SkillEntity, now2 int64) {
	cfg := entity.Cfg
	phase := entity.MovePhase
	move := cfg.SkillMove[phase]
	dist := move.Dist
	dir := entity.SkillPos.Dir
	mi := entity.MapInfo
	pos := mi.GetPos()
	// 计算新的坐标
	x, y := CalcNewWalkAbleXY(state, pos, dir, dist)
	// 理论上前端会采用相同算法，所以不需要通知
	oldX, oldY := pos.X, pos.Y
	pos.X = x
	pos.Y = y
	pos.Dir = dir
	// 更新视野
	AoiUpdatePos(state, oldX, oldY, pos, mi)
	if cfg.MoveChangePos == 1 {
		// 更新新的施法坐标
		skillPos := entity.SkillPos
		skillPos.X,skillPos.Y = x,y
	}
	phase++
	if int(phase) < len(cfg.SkillMove) {
		entity.MovePhase = phase
		entity.MoveTime += int64(cfg.SkillMove[phase].StartTime)
	} else {
		entity.MoveTime = 0
	}
}

var SkillSearchPK AreaFoldFunc = func(state *MapState, srcMI, destMI MapInfo, args ...unsafe.Pointer) bool {
	if !destMI.IsAlive() {
		return false
	}
	entity := (*SkillEntity)(args[0])
	if flyInfo := entity.FlyInfo; flyInfo != nil {
		if flyInfo.IsEffected(destMI) {
			return false
		}
	}
	srcType := srcMI.Type()
	destType := destMI.Type()
	switch srcType {
	case global.ACTOR_ROLE:
		if destType == global.ACTOR_ROLE {
			// 判断pk模式
			return true
		} else {
			// 判断一下boss次数之类
			return true
		}
	case global.ACTOR_MONSTER:
		// 通常是判断一下阵营
		return true
	}
	return true
}

// 不太可能需要热更
func CalcEffect(state *MapState, entity *SkillEntity, srcActor *MapActor, effectID int32, targets []MapInfo,
	effectList []*global.PSkillEffect) []*global.PSkillEffect {
	effect := config.SkillEffect.Get(effectID)
	f := SkillEffectMap[effect.Func]
	for _, target := range targets {
		effectList = (*f)(state, entity, srcActor, target, targets, effectList, effect.Arg)
	}
	return effectList
}

func EffectToAreas(state *MapState, targets []MapInfo) map[Area]bool {
	ms := make(map[Area]bool, 12)
	for _, e := range targets {
		pos := state.GetMapInfo(e.Type(), e.ID()).GetPos()
		l := Get9AreaByPos(pos)
		for _, area := range l {
			if _, ok := ms[area]; !ok {
				ms[area] = true
			}
		}
	}
	return ms
}

// 计算新的施法位置，大概率不需要热更
func CalcCurrentParam(entity *SkillEntity, effectShape *config.SkillShape, effectSpread,
	phase int32) (pos *global.PPos, shape int32, width, height float64) {
	var deviation float32
	pos = entity.SkillPos
	switch effectShape.ShapeType {
	case SHAPE_TYPE_RECTANGLE:
		deviation = effectShape.A
		width = float64(effectShape.B)
		height = float64(effectShape.C)
		// 只有矩形是飞行子弹
		if effectSpread == SPREAD_FLY {
			fly := entity.FlyInfo
			if effectShape.D != 0 {
				if fly.Frames > 0 {
					width = math.Pow(float64(effectShape.D), float64(fly.Frames)) * width
				}
			}
			// 计算偏移
			cfg := entity.Cfg.SkillSpreadSpeed[phase]
			speed := cfg.Key
			deviation += float32(speed * FRAMES_TIME / 1000)
		}
	case SHAPE_TYPE_CIRCLE:
		deviation = effectShape.B
		width = float64(effectShape.A)
	case SHAPE_TYPE_SECTOR: // 扇形通常是没有偏移的
		deviation = 0
		width = float64(effectShape.A)  // 半径
		height = float64(effectShape.B) // 夹角大小
	}
	if deviation > 0 {
		dir := pos.Dir
		pos.X += float32(lib.Cos(dir) * float64(deviation))
		pos.Y += float32(lib.Sin(dir) * float64(deviation))
	}
	return
}

func MakeSkillEntityID(state *MapState) int32 {
	data := state.DataSkill
	data.Index++
	return data.Index
}

// 这里把伤害效果和回调分开两个函数处理
var ActiveCallBack = func(state *MapState, entity *SkillEntity, effectList []int32, targets []MapInfo) {
	for _, effectID := range effectList {
		effectCfg := config.SkillEffect.Get(effectID)
		if ef, ok := SkillCallBackMap[effectCfg.Func]; ok {
			(*ef)(state, entity, targets, effectCfg.Arg)
		}
	}
}

// 统一释放资源到池
func (S *SkillEntity) Release(state *MapState) {
	actor := state.GetActorByID(S.MapInfo.Type(), S.MapInfo.ID())
	for i, v := range actor.Skills {
		if v == S.ID {
			// 这里可以在迭代器里面执行
			endIdx := len(actor.Skills) - 1
			actor.Skills[i] = actor.Skills[endIdx]
			actor.Skills = actor.Skills[0:endIdx]
			break
		}
	}
	_skillPosPool.Put(S.InitPos)
	_skillPosPool.Put(S.SkillPos)
	if S.FlyInfo != nil {
		_skillFlyPool.Put(S.FlyInfo)
		S.FlyInfo = nil
	}
	S.Target = nil
	S.Cfg = nil
	S.MapInfo = nil
	_skillEntityPool.Put(S)
}

//-------------------------------------------------------
func (s *SkillFlyInfo) IsEffected(mi MapInfo) bool {
	_, ok := s.EffectActors[AKey{mi.ID(), mi.Type()}]
	return ok
}
