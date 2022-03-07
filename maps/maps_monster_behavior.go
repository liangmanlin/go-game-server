package maps

import (
	"game/config"
	"game/global"
	"game/lib"
	"github.com/liangmanlin/gootp/astar"
	"github.com/liangmanlin/gootp/kernel"
	"sync"
	"unsafe"
)

const (
	BTYPE_FIGHT int32 = iota + 1
	BTYPE_COND        // 条件
	BTYPE_FUNC        // 直接执行
	BTYPE_LIST
)

const MONSTER_WORK_TICK int64 = 200 // ms

var _enemyPool = sync.Pool{
	New: func() interface{} {
		return &MonsterEnemy{}
	},
}

func NewEnemy() *MonsterEnemy {
	return _enemyPool.Get().(*MonsterEnemy)
}

var _doingPool = sync.Pool{
	New: func() interface{} {
		return &Doing{}
	},
}

type behaviorFunc = func(state *MapState, mState *MState, now2 int64, args interface{}) bool

var _behaviorFun = make(map[string]*behaviorFunc, 100)

func NewDoing() *Doing {
	doing := _doingPool.Get().(*Doing)
	return doing
}

func DelDoing(mState *MState) {
	size := len(mState.Doings)
	if size > 0 {
		size --
		doing := mState.Doings[size]
		mState.Doings[size] = nil // gc
		_doingPool.Put(doing)
		mState.Doings = mState.Doings[0 : size]
	}
}

func BehaviorActTree(state *MapState, mState *MState, now2 int64) {
	cfg := config.MonsterBehavior.Get(mState.Config.Behavior)
	switch cfg.Type {
	case BTYPE_COND:
		f := GetBehaviorFun(cfg.Func)
		if (*f)(state, mState, now2, cfg.Arg) {
			BehaviorActChild(state, mState, now2, cfg)
		}
	case BTYPE_FUNC:
		mState.AddDoing(cfg.Type, GetBehaviorFun(cfg.Func), cfg.Arg)
	}
}

func BehaviorActChild(state *MapState, mState *MState, now2 int64, cfg *config.DefMonsterBehavior) {
	child := cfg.Child
	for _, bid := range child {
		cfg = config.MonsterBehavior.Get(bid)
		switch cfg.Type {
		case BTYPE_COND:
			f := GetBehaviorFun(cfg.Func)
			if (*f)(state, mState, now2, cfg.Arg) {
				BehaviorActChild(state, mState, now2, cfg)
			}
		default:
			mState.AddDoing(cfg.Type, GetBehaviorFun(cfg.Func), cfg.Arg)
			break
		}
	}
}

var MoveToPos = func(state *MapState, mState *MState, now2 int64, dx, dy float32, idleTime int32) {
	pos := state.MapMonsters[mState.ID].Pos
	rt, movePath := AStarSearch(state, pos.X, pos.Y, dx, dy)
	switch rt {
	case astar.ASTAR_SAME_POS:
		mState.NextTime = now2 + int64(idleTime)
	case astar.ASTAR_FOUNDED:
		MoveStart(state, global.ACTOR_MONSTER, mState.ID, pos.X, pos.Y, movePath)
		mState.DelDoing(&MovePos)
		mState.AddDoing(BTYPE_FUNC, &MovePos, &MoveArg{dx, dy, idleTime})
	case astar.ASTAR_UNREACHABLE:
		kernel.ErrorLog("mapID:%d,cannot move from {%f,%f} to {%f,%f}", state.Config.MapID, pos.X, pos.Y, dx, dy)
		ReturnBorn(state, mState)
	}
}

func ReturnBorn(state *MapState, mState *MState) {
	mi := state.MapMonsters[mState.ID]
	pos := mi.Pos
	if lib.XYLessThan(pos.X, pos.Y, mState.BornPos.X, mState.BornPos.Y, 4) {
		MoveMonsterStop(state, mState.ID, true)
		ReturnBornFinish(state, mState)
	} else {
		// 满血
		if mi.HP != mi.MaxHP {
			mi.HP = mi.MaxHP
			// 广播通知
			state.UpdateMonsterInfo(mState.ID, pos, []*global.PKV{{MAP_MONSTER_HP, mi.HP}})
		}
		MoveBornPos(state, mState, mState.BornPos)
	}
}

var MoveBornPos = func(state *MapState, mState *MState, destPos *global.PPos) {
	pos := state.MapMonsters[mState.ID].Pos
	rt, movePath := AStarSearch(state, pos.X, pos.Y, destPos.X, destPos.Y)
	switch rt {
	case astar.ASTAR_SAME_POS:
		ReturnBornFinish(state, mState)
	case astar.ASTAR_FOUNDED:
		MoveStart(state, global.ACTOR_MONSTER, mState.ID, pos.X, pos.Y, movePath)
		mState.CleanDoing() // 清空所有
		mState.AddDoingName(BTYPE_FUNC, BF_RETURN_BORN_POS, destPos)
	case astar.ASTAR_UNREACHABLE:
		kernel.ErrorLog("mapID:%d,cannot move from {%f,%f} to {%f,%f}", state.Config.MapID, pos.X, pos.Y, destPos.X, destPos.Y)
		ReturnBornFinish(state, mState)
	}
}

func ReturnBornFinish(state *MapState, mState *MState) {
	// 在判断一次血量
	mi := state.MapMonsters[mState.ID]
	if mi.HP != mi.MaxHP {
		mi.HP = mi.MaxHP
		// 广播通知
		state.UpdateMonsterInfo(mState.ID, mi.Pos, []*global.PKV{{MAP_MONSTER_HP, mi.HP}})
	}
	// 删除所有事件
	mState.CleanDoing()
	// 清空enemy
	mState.CleanEnemy()
}

var GetMonsterEnemy = func(state *MapState, mState *MState) (*MonsterEnemy, bool) {
	if len(mState.Enemies) == 0 {
		return nil, false
	}
	bornPos := mState.BornPos
	activeRadius := float32(mState.Config.ActivityRadius)
	var enemy *MonsterEnemy
	for _, e := range mState.Enemies {
		mi := state.GetMapInfo(e.ActorType, e.ActorID)
		if mi == nil {
			continue
		}
		pos := mi.GetPos()
		if mi.IsAlive() && lib.PosLessThan(bornPos, pos, activeRadius) {
			if enemy == nil {
				enemy = e
			} else {
				if enemy.TotalHurt < e.TotalHurt {
					enemy = e
				}
			}
		}
	}
	return enemy, enemy != nil
}

var MoveToDestPath = func(state *MapState, mState *MState, monster *global.PMapMonster, dMi MapInfo, dist int32, now2 int64) {
	dPos := dMi.GetPos()
	pos := monster.Pos
	rt, movePath := AStarSearch(state, pos.X, pos.Y, dPos.X, dPos.Y)
	switch rt {
	case astar.ASTAR_SAME_POS:
	case astar.ASTAR_FOUNDED:
		MoveStart(state, global.ACTOR_MONSTER, mState.ID, pos.X, pos.Y, movePath)
		fPtr := GetBehaviorFun(BF_MOVE_DEST)
		mState.DelDoing(fPtr)
		mState.AddDoing(BTYPE_FUNC, fPtr, &MoveDestArg{dMi.ID(), dMi.Type(), dist, dPos.X, dPos.Y})
	case astar.ASTAR_UNREACHABLE:
		kernel.ErrorLog("mapID:%d,cannot move from {%f,%f} to {%f,%f}", state.Config.MapID, pos.X, pos.Y, dPos.X, dPos.Y)
		ReturnBorn(state, mState)
	}
}

func RemoveToDestPath(state *MapState, mState *MState, mi MapInfo, dist int32, actorType int8, actorID, now2 int64) {
	bornPos := mState.BornPos
	activeRadius := mState.Config.ActivityRadius
	if lib.PosLessThan(bornPos, mi.GetPos(), float32(activeRadius)) {
		MoveToDestPath(state, mState, state.MapMonsters[mState.ID], mi, dist, now2)
	} else {
		DelEnemy(state, mState, actorType, actorID, now2)
	}
}

func DelEnemy(state *MapState, mState *MState, actorType int8, actorID, now2 int64) {
	aKey := AKey{actorID, actorType}
	delete(mState.Enemies, aKey)
	if len(mState.Enemies) == 0 {
		ReturnBorn(state, mState)
	} else {
		mState.DelDoingName(BF_MOVE_DEST)
	}
}

// 尝试环绕目标，这里是先插入一个列表，再后续处理
// 因为可能下一刻，同一个怪也会走到同一个坐标
func TryAroundDest(state *MapState, pos *global.PPos, monsterID int64, actorType int8, actorID int64) {
	if state.PosHaveMonster(pos) {
		aKey := AKey{ActorID: actorID, ActorType: actorType}
		dm := state.DataMonster
		if list, ok := dm.AroundDest[aKey]; ok {
			dm.AroundDest[aKey] = append(list, monsterID)
		} else {
			dm.AroundDest[aKey] = []int64{monsterID}
		}
	}
}

var BReleaseSkill = func(state *MapState, mState *MState, now2 int64, destMI MapInfo) {
	// 获取可以释放的技能
	var skillID int32
	for i := range mState.Skills {
		skill := mState.Skills[i]
		if now2 >= skill.UseTime {
			skillID = skill.SkillID
			break
		}
	}
	if skillID > 0 {
		cfg := config.Skill.Get(skillID)
		mi := state.MapMonsters[mState.ID]
		dPos := destMI.GetPos()
		dir := lib.CalcDir(mi.Pos, dPos)
		var target *global.PActor
		if cfg.PosType == SKILL_POS_TYPE_3 || cfg.PosType == SKILL_POS_TYPE_4 {
			target = &global.PActor{ActorType: destMI.Type(), ActorID: destMI.ID()}
		}
		MonsterReleaseSKill(state, mi, skillID, dPos.X, dPos.Y, int16(dir), target, now2)
	}
}

var SearchEnemy AreaFoldFunc = func(state *MapState, srcInfo, destInfo MapInfo, args ...unsafe.Pointer) bool {
	dist := *(*int32)(args[0])
	camp := *(*int8)(args[1])
	if lib.PosLessThan(srcInfo.GetPos(), destInfo.GetPos(), float32(dist)) && destInfo.IsAlive() &&
		(camp == 0 || destInfo.GetCamp() != camp) {
		return true
	}
	return false
}

func BuildEnemies(mState *MState, ml []MapInfo, now2 int64) {
	for i := range ml {
		dmi := ml[i]
		enemy := NewEnemy()
		enemy.ActorID = dmi.ID()
		enemy.ActorType = dmi.Type()
		enemy.TotalHurt = 0
		enemy.LastAttackTime = now2
		mState.Enemies[AKey{ActorType: enemy.ActorType, ActorID: enemy.ActorID}] = enemy
	}
}

// -----------------------------------------------------------------------------------------

// 通用战斗节点
const BF_FIGHT = "fight"

var Fight behaviorFunc = func(state *MapState, mState *MState, now2 int64, args interface{}) bool {
	if mState.LastAttackTime+int64(mState.Config.AttackSpeed) > now2 {
		return false
	}
	enemy, ok := GetMonsterEnemy(state, mState)
	if !ok {
		// 没有目标
		rType, time := mState.ReturnType()
		switch rType {
		case RETURN_TYPE_TRUE:
			ReturnBorn(state, mState)
		case RETURN_TYPE_FALSE:
			mState.NextTime = now2 + int64(mState.Config.AttackSpeed)
		case RETURN_TYPE_TIME:
			if now2 >= mState.LastAttackTime+time {
				ReturnBorn(state, mState)
			} else {
				mState.NextTime = now2 + int64(mState.Config.AttackSpeed)
			}
		}
		return false
	}
	// 有目标，发起攻击
	monster := state.MapMonsters[mState.ID]
	dMi := state.GetMapInfo(enemy.ActorType, enemy.ActorID)
	dPos := dMi.GetPos()
	if lib.PosLessThan(monster.Pos, dPos, float32(mState.Config.AttackRadius)) {
		// 距离足够，可以施法
		BReleaseSkill(state, mState, now2, dMi)
	} else {
		// 移动到目标点
		MoveToDestPath(state, mState, monster, dMi, mState.Config.AttackRadius, now2)
	}
	return true
}

const BF_PATROL = "patrol"

var Patrol behaviorFunc = func(state *MapState, mState *MState, now2 int64, args interface{}) bool {
	if state.Rand.Int32(100) <= 20 {
		DelDoing(mState)
		// 计算一个随机点
		if x, y, ok := MakeRandomXY(state, mState.BornPos, 4); ok {
			idleTime := args.([]int32)[0] // 毫秒
			MoveToPos(state, mState, now2, x, y, idleTime)
		} else {
			mState.NextTime = now2 + MONSTER_WORK_TICK*2
		}
	}
	return true
}

const BF_MOVE_POS = "move_pos"

var MovePos behaviorFunc = func(state *MapState, mState *MState, now2 int64, args interface{}) bool {
	if state.MonsterActor(mState.ID).IsMove {
		mState.NextTime = now2 + MONSTER_WORK_TICK*2
	} else {
		// 移动到目标点，或者被打断了移动
		DelDoing(mState)
		mState.NextTime = now2 + int64(args.(*MoveArg).IdleTime)
	}
	return true
}

const BF_MOVE_DEST = "move_dest"

var MoveDest behaviorFunc = func(state *MapState, mState *MState, now2 int64, args interface{}) bool {
	arg := args.(*MoveDestArg)
	// 确认目标是否存在
	mi := state.GetMapInfo(arg.ActorType, arg.ActorID)
	if mi == nil {
		DelDoing(mState)
		return false
	}
	miPos := mi.GetPos()
	pos := state.MapMonsters[mState.ID].Pos
	if lib.PosLessThan(pos, miPos, float32(arg.Dist)) {
		MoveMonsterStop(state, mState.ID, true)
		// 尝试围绕目标
		TryAroundDest(state, pos, mState.ID, arg.ActorType, arg.ActorID)
	} else {
		actor := state.GetActorByID(global.ACTOR_MONSTER, mState.ID)
		if actor.IsMove {
			if !lib.XYLessThan(miPos.X, miPos.Y, arg.DX, arg.DY, 3) {
				// 变化比较大，从新移动
				RemoveToDestPath(state, mState, mi, arg.Dist, arg.ActorType, arg.ActorID, now2)
			}
		} else {
			// 走到这个分支，肯定是距离不够的
			RemoveToDestPath(state, mState, mi, arg.Dist, arg.ActorType, arg.ActorID, now2)
		}
	}
	return true
}

const BF_RETURN_BORN_POS = "return_born_pos"

var ReturnBornPos behaviorFunc = func(state *MapState, mState *MState, now2 int64, args interface{}) bool {
	if !state.MonsterActor(mState.ID).IsMove {
		pos := state.MapMonsters[mState.ID].Pos
		destPos := args.(*global.PPos)
		if lib.XYLessThan(pos.X, pos.Y, destPos.X, destPos.Y, 4) {
			ReturnBornFinish(state, mState)
		} else {
			MoveBornPos(state, mState, destPos)
		}
	}
	return true
}

const BF_SEARCH_ENEMY_NEIGHBOR = "search_enemy_neighbor"

var SearchEnemyNeighbor behaviorFunc = func(state *MapState, mState *MState, now2 int64, args interface{}) bool {
	arg := args.([]int32)
	mi := state.MapMonsters[mState.ID]
	areaList := Get9AreaByPos(mi.Pos)
	dist := mState.Config.GuardRadius
	var camp int8
	if arg[1] == 1 {
		camp = mi.Camp
	}
	if t := arg[0]; t == 0 {
		// 同时攻击人和怪物
		findRole := AreasActorFold(state, global.ACTOR_ROLE, areaList, mi, SearchEnemy, unsafe.Pointer(&dist), unsafe.Pointer(&camp))
		BuildEnemies(mState, findRole, now2)
		ReleaseMapInfoSlice(findRole)
		findMonster := AreasActorFold(state, global.ACTOR_MONSTER, areaList, mi, SearchEnemy, unsafe.Pointer(&dist), unsafe.Pointer(&camp))
		BuildEnemies(mState, findMonster, now2)
		ReleaseMapInfoSlice(findMonster)
	} else {
		find := AreasActorFold(state, int8(t), areaList, mi, SearchEnemy, unsafe.Pointer(&dist), unsafe.Pointer(&camp))
		BuildEnemies(mState, find, now2)
		ReleaseMapInfoSlice(find)
	}
	if len(mState.Enemies) > 0 {
		return true
	}
	return false
}

func init() {
	_behaviorFun[BF_PATROL] = &Patrol
	_behaviorFun[BF_MOVE_POS] = &MovePos
	_behaviorFun[BF_RETURN_BORN_POS] = &ReturnBornPos
	_behaviorFun[BF_FIGHT] = &Fight
	_behaviorFun[BF_MOVE_DEST] = &MoveDest
	_behaviorFun[BF_SEARCH_ENEMY_NEIGHBOR] = &SearchEnemyNeighbor
}

func GetBehaviorFun(funName string) *behaviorFunc {
	return _behaviorFun[funName]
}
