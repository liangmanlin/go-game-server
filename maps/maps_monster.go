package maps

import (
	"game/config"
	"game/global"
	"game/lib"
	"github.com/liangmanlin/gootp/gutil"
	"github.com/liangmanlin/gootp/kernel"
	"math"
	"sync"
)

var _argPool = sync.Pool{
	New: func() interface{} {
		return &MonsterCreateArg{}
	},
}
var CreateMonsterArg = func() *MonsterCreateArg {
	arg := _argPool.Get().(*MonsterCreateArg)
	arg.Dir = -1
	arg.Level = 0
	arg.DeadArgs = nil
	arg.DeadCB = nil
	arg.Prop = nil
	return arg
}

// 不太可能热更
func MonsterLoop(state *MapState, now2 int64) {
	dm := state.DataMonster
	if dm.StopMonster == 0 || dm.StopMonster > now2 {
		dm.IsMonsterLoop = true
		for _, mState := range state.Monsters {
			monsterUpdate(state, mState, now2)
		}
		dm.IsMonsterLoop = false
		if len(dm.DelMonster) > 0 {
			for _, id := range dm.DelMonster {
				MonsterDelete(state, state.Monsters[id])
			}
			dm.DelMonster = dm.DelMonster[0:0]
		}
	}
}

// 不太可能热更
func MonsterUpdateSecond(state *MapState, now2 int64) {
	dm := state.DataMonster
	if len(dm.AroundDest) > 0 {
		// 处理环绕
		for ak, l := range dm.AroundDest {
			actor := state.GetActor(ak)
			if actor == nil || actor.State == ACTOR_STATE_DEAD {
				continue
			}
			mi := state.GetMapInfoAKey(ak)
			for _, monsterID := range l {
				monsterInfo := state.GetMapInfoMonster(monsterID)
				if monsterInfo == nil || monsterInfo.HP == 0 || !state.PosHaveMonster(monsterInfo.Pos) {
					continue
				}
				mState := state.Monsters[monsterID]
				attackRadius := mState.Config.AttackRadius
				if x, y, ok := GetFreeAroundPos(state, mi.GetPos(), monsterInfo.Pos, float32(attackRadius)-0.5); ok {
					MoveToPos(state, mState, now2, x, y, 0)
				}
			}
			delete(dm.AroundDest, ak)
		}
	}
}

// 计算一个环绕坐标，如果所有目标都有怪物，就采用当前位置
func GetFreeAroundPos(state *MapState, centerPos, pos *global.PPos, dist float32) (x, y float32, ok bool) {
	// 计算起始位置的角度
	// 假设同一面的怪物是聚集点，所以随机一个300度
	for i := 0; i < 20; i++ {
		dir := lib.CalcDir(centerPos, pos)
		dir = state.Rand.Random(0, 300) + dir + 30
		r := lib.DirToTan(dir)
		x = centerPos.X + float32(math.Cos(r))*dist
		y = centerPos.Y + float32(math.Sin(r))*dist
		if state.XYWalkAble(x, y) {
			ok = true
			break
		}
	}
	return
}

func CreateMapAllMonster(state *MapState) {
	cfg := state.Config
	for i := range cfg.Monsters {
		m := &cfg.Monsters[i]
		arg := CreateMonsterArg()
		arg.X = float32(m.X)
		arg.Y = float32(m.Y)
		arg.TypeID = m.TypeID
		CreateMonster(state, arg)
	}
}

var CreateMonster = func(state *MapState, arg *MonsterCreateArg) (monsterID int64) {
	// 最后归还
	defer _argPool.Put(arg)
	cfg := config.Monster.Get(arg.TypeID)
	if cfg == nil {
		kernel.ErrorLog("monster %d config not found", arg.TypeID)
		return
	}
	monsterID = state.MakeMonsterID()
	bornPos := &global.PPos{X: arg.X, Y: arg.Y, Dir: arg.Dir}
	mState := &MState{ID: monsterID, TypeID: arg.TypeID,
		BornPos: bornPos, DeadCB: arg.DeadCB, DeadArgs: arg.DeadArgs,
		Enemies: make(map[AKey]*MonsterEnemy),
	}
	RebornMonster(state, mState, arg.Level, cfg, arg.Prop)
	state.Monsters[monsterID] = mState
	return
}

var RebornMonster = func(state *MapState, mState *MState, level int16, cfg *config.DefMonster, prop *global.PProp) {
	mState.Config = cfg
	if prop == nil {
		prop = lib.ToProp(config.MonsterProp.Get(cfg.PropID))
	}
	if level == 0 {
		level = int16(cfg.Level)
	}
	var pos *global.PPos
	var mapInfo *global.PMapMonster
	if t, ok := state.MapMonsters[mState.ID]; ok {
		mapInfo = t
		mapInfo.HP = prop.MaxHP
		mapInfo.MaxHP = prop.MaxHP
		mapInfo.MoveSpeed = prop.MoveSpeed
		mapInfo.Level = level
		mapInfo.Pos.X = mState.BornPos.X
		mapInfo.Pos.Y = mState.BornPos.Y
		mapInfo.Pos.Dir = mState.BornPos.Dir
		pos = mapInfo.Pos
	} else {
		p := *mState.BornPos
		pos = &p
		mapInfo = &global.PMapMonster{MonsterID: mState.ID, TypeID: mState.TypeID, HP: prop.MaxHP, MaxHP: prop.MaxHP,
			Level: level, Pos: pos, MoveSpeed: prop.MoveSpeed,
		}
	}
	if pos.Dir < 0 {
		pos.Dir = int16(state.Rand.Random(0, 359))
	}
	aKey := AKey{ActorID: mState.ID, ActorType: global.ACTOR_MONSTER}
	actor := state.GetActor(aKey)
	if actor == nil {
		actor = NewActor(global.ACTOR_MONSTER, mState.ID)
		state.Actors[aKey] = actor
	}
	actor.State = ACTOR_STATE_NORMAL
	mState.State = ACTOR_STATE_NORMAL
	state.MapMonsters[mState.ID] = mapInfo
	state.Monsters[mState.ID] = mState
	if len(mState.Skills) == 0 {
		mState.Skills = MakeMonsterSkill(cfg)
	}
	grid := GetGridByPos(pos)
	state.EnterActor(GetAreaByGrid(grid), global.ACTOR_MONSTER, mState.ID)
	state.PosMonsterUP(grid, 1)
	// 通知视野的人
	proto := &global.MapTocMonsterEnterArea{MapInfo: mapInfo}
	state.BroadCastPos(pos, proto)
}

func MakeMonsterSkill(cfg *config.DefMonster) []MonsterSkill {
	var rs []MonsterSkill
	for _, skillID := range cfg.Skills {
		cd := config.Skill.Get(skillID).SkillCD
		rs = append(rs, MonsterSkill{SkillID: skillID, CD: cd})
	}
	return rs
}

func monsterUpdate(state *MapState, mState *MState, now2 int64) {
	if now2 >= mState.NextTime {
		MonsterUpdate(state, mState, now2)
	}
}

var MonsterUpdate = func(state *MapState, mState *MState, now2 int64) {
	if mState.State == ACTOR_STATE_DEAD {
		cfg := config.Monster.Get(mState.TypeID)
		if cfg.RefreshTime > 0 {
			level := state.GetMapInfoMonster(mState.ID).Level
			RebornMonster(state, mState, level, cfg, state.Actors[AKey{ActorType: global.ACTOR_MONSTER, ActorID: mState.ID}].BaseProp)
		} else {
			MonsterDelete(state, mState)
		}
		return
	}
	size := len(mState.Doings)
	if size == 0 {
		BehaviorActTree(state, mState, now2)
	} else {
		do := mState.Doings[size-1]
		(*do.Fun)(state, mState, now2, do.Args)
	}
}

func MonsterDelete(state *MapState, mState *MState) {
	dm := state.DataMonster
	if dm.IsMonsterLoop {
		dm.DelMonster = append(dm.DelMonster, mState.ID)
		return
	}
	aKey := AKey{ActorType: global.ACTOR_MONSTER, ActorID: mState.ID}
	if mState.State != ACTOR_STATE_DEAD {
		MoveMonsterStop(state, mState.ID, false)
		mapInfo := state.GetMapInfoMonster(mState.ID)
		grid := GetGridByPos(mapInfo.Pos)
		state.LeaveActor(GetAreaByGrid(grid), global.ACTOR_MONSTER, mState.ID)
		state.PosMonsterUP(grid, -1)
		state.BroadCastPos(mapInfo.Pos, &global.MapTocActorLeaveArea{ActorID: mState.ID, ActorType: global.ACTOR_MONSTER})
	}
	delete(state.Monsters, mState.ID)
	delete(state.MapMonsters, mState.ID)
	delete(state.Actors, aKey)
}

var MonsterReduceHP = func(state *MapState, srcActor, destActor *MapActor, destMI *global.PMapMonster, damage int32) {
	mState := state.Monsters[destMI.MonsterID]
	AddEnemy(state, mState, srcActor, damage)
	if destMI.HP > damage {
		destMI.HP -= damage
	} else {
		MonsterDead(state, mState, srcActor, destActor, destMI)
		// 执行死亡后在设置血量为0
		destMI.HP = 0
	}
}

var MonsterDead = func(state *MapState, mState *MState, srcActor, destActor *MapActor, destMI *global.PMapMonster) {
	monsterID := destMI.MonsterID
	MoveMonsterStop(state, monsterID, false)
	destActor.State = ACTOR_STATE_DEAD
	mState.State = ACTOR_STATE_DEAD
	grid := GetGridByPos(destMI.Pos)
	state.LeaveActor(GetAreaByGrid(grid), global.ACTOR_MONSTER, monsterID)
	state.PosMonsterUP(grid, -1)
	mState.NextTime = kernel.Now2() + int64(mState.Config.RefreshTime)
	proto := &global.MapTocActorDead{ActorID: monsterID, ActorType: global.ACTOR_MONSTER}
	state.BroadCastPos(destMI.Pos, proto)
	if mState.DeadCB != nil {
		(*mState.DeadCB)(state, monsterID, srcActor, mState.DeadArgs)
	}
}

func AddEnemy(state *MapState, mState *MState, srcActor *MapActor, damage int32) {
	key := AKey{ActorID: srcActor.ActorID, ActorType: srcActor.ActorType}
	if enemy, ok := mState.Enemies[key]; ok {
		enemy.LastAttackTime = kernel.Now2()
		enemy.TotalHurt += int64(damage)
	} else {
		enemy = NewEnemy()
		enemy.TotalHurt = int64(damage)
		enemy.LastAttackTime = kernel.Now2()
		enemy.ActorType = srcActor.ActorType
		enemy.ActorID = srcActor.ActorID
		mState.Enemies[key] = enemy
		// 如果第一次被攻击，需要切换到战斗状态
		if len(mState.Enemies) == 1 {
			TryFight(mState)
		}
	}
}

func TryFight(mState *MState)  {

}

func MonsterUpdateHPMoveSpeed(state *MapState,actor *MapActor,upList []global.PKV) {
	MI := state.MapMonsters[actor.ActorID]
	ul := make([]*global.PKV,0,len(upList))
	for i := range upList {
		v := &upList[i]
		switch v.Key {
		case lib.PROP_MaxHP:
			if MI.HP >= v.Value {
				MI.HP = v.Value
			}else{
				MI.HP = gutil.Trunc(float32(MI.HP)/float32(MI.MaxHP)*float32(v.Value))
			}
			MI.MaxHP = v.Value
			v.Key = MAP_MONSTER_MAX_HP
			ul = append(ul,v,&global.PKV{Key: MAP_MONSTER_MAX_HP,Value: MI.HP})
		case lib.PROP_MoveSpeed:
			if v.Value != MI.MoveSpeed {
				MI.MoveSpeed = v.Value
				v.Key = MAP_MONSTER_MOVE_SPEED
				ul = append(ul,v)
			}
		}
	}
	state.UpdateMonsterInfo(MI.MonsterID,MI.Pos,ul)
}

func (m *MState) DelDoing(funPtr *behaviorFunc) {
	size := len(m.Doings)
	if size > 0 {
		doing := m.Doings[size-1]
		if doing.Fun == funPtr {
			_doingPool.Put(doing)
			m.Doings = m.Doings[0 : size-1]
		}
	}
}

func (m *MState) DelDoingName(funName string) {
	ptr := GetBehaviorFun(funName)
	m.DelDoing(ptr)
}

func (m *MState) CleanDoing() {
	size := len(m.Doings)
	if size > 0 {
		for i := 0; i < size; i++ {
			doing := m.Doings[i]
			_doingPool.Put(doing)
		}
		m.Doings = m.Doings[0:0]
	}
}

func (m *MState) AddDoing(bType int32, fun *behaviorFunc, args interface{}) {
	doing := NewDoing()
	doing.Type = bType
	doing.Fun = fun
	doing.Args = args
	m.Doings = append(m.Doings, doing)
}

func (m *MState) AddDoingName(bType int32, funName string, args interface{}) {
	m.AddDoing(bType, GetBehaviorFun(funName), args)
}

func (m *MState) ReturnType() (ReturnType, int64) {
	if m.Return != nil {
		return m.Return.Type, m.Return.Time
	}
	return RETURN_TYPE_TRUE, 0
}

func (m *MState) CleanEnemy() {
	size := len(m.Enemies)
	if size > 0 {
		for k, v := range m.Enemies {
			_enemyPool.Put(v)
			delete(m.Enemies, k)
		}
	}
}

