package maps

import (
	"game/global"
	"github.com/liangmanlin/gootp/kernel"
	"math"
	"sync"
)

const DELTA_ERROR float32 = 64
const DIST_FAIL float32 = 25

var _movePool = sync.Pool{
	New: func() interface{} {
		return &Move{}
	},
}

func MakeMove() *Move {
	return _movePool.Get().(*Move)
}

// 开始移动
var MoveStart = func(state *MapState, actorType int8, actorID int64, startX, startY float32, movePath *global.PMovePath) {
	now2 := kernel.Now2()
	aKey := AKey{ActorID: actorID, ActorType: actorType}
	actor := state.GetActor(aKey)
	if actorType == global.ACTOR_ROLE && actor.IsMove {
		move := state.MoveActor[aKey]
		if move != nil {
			ActorMove(state, actor, move, now2, false)
		}
	}
	mapInfo := state.GetMapInfo(actorType, actorID)
	pos := mapInfo.GetPos()
	// 先简单判断一下
	if actorType == global.ACTOR_MONSTER ||
		(MoveCheckDist(state, actorType, actorID, pos, startX, startY) && state.XYWalkAble(startX, startY)) {
		var dir int16
		move := MakeMove()
		move.LastUpdateTime = now2
		move.StepCount = 0
		if len(movePath.GridPath) == 0 {
			move.DestX = movePath.EndX
			move.DestY = movePath.EndY
		} else {
			grid := movePath.GridPath[0]
			move.DestX = float32(grid.GridX)
			move.DestY = float32(grid.GridY)
		}
		_, _, dir = CalcNewXY(startX, startY, move.DestX, move.DestY, 0)
		move.StepTotal = CalcTotalMove(startX, startY, move.DestX, move.DestY)
		state.MoveActor[aKey] = move
		actor.IsMove = true
		oldX := pos.X
		oldY := pos.Y
		pos.X = startX
		pos.Y = startY
		pos.Dir = dir
		mapInfo.SetMovePath(movePath)
		// 可能起始位置有变化，所以这里需要刷新一下视野
		if actorType == global.ACTOR_ROLE {
			AoiUpdatePos(state, oldX, oldY, pos, mapInfo)
		}
		proto := &global.MapTocMove{ActorType: actorType, ActorID: actorID, StartX: startX, StartY: startY, MovePath: movePath}
		//kernel.ErrorLog("%s",lib.ToString(proto))
		state.BroadCastPos(pos, proto)
		return
	}
	// 说明不合法，需要停止移动
	kernel.ErrorLog("un walkable:%d,%d,%#v,%f,%f,%v", actorID, now2, pos, startX, startY, actor.IsMove)
	SyncStop(state, actor, true)
}

// 角色停止移动
var MoveRoleStop = func(state *MapState, roleID int64, stopX, stopY float32, dir int16) {
	mapInfo := state.MapRoles[roleID]
	if mapInfo == nil {
		return
	}
	aKey := AKey{ActorType: global.ACTOR_ROLE, ActorID: roleID}
	actor := state.GetActor(aKey)
	if actor.IsMove {
		DelMove(state.MoveActor, aKey)
		actor.IsMove = false
		mapInfo.MovePath = nil
		if MoveCheckDist(state, global.ACTOR_ROLE, roleID, mapInfo.Pos, stopX, stopY) &&
			state.XYWalkAble(stopX, stopY) {
			oldX := mapInfo.Pos.X
			oldY := mapInfo.Pos.Y
			mapInfo.Pos.X = stopX
			mapInfo.Pos.Y = stopY
			mapInfo.Pos.Dir = dir
			AoiUpdatePos(state, oldX, oldY, mapInfo.Pos, mapInfo)
			proto := &global.MapTocMoveStop{ActorType: global.ACTOR_ROLE, ActorID: roleID, X: stopX, Y: stopY, Dir: dir}
			state.BroadCastPosExclude(mapInfo.Pos, roleID, proto)
		} else {
			SyncStop(state, actor, false)
		}
	} else {
		// 没有在移动
		proto := &global.MapTocStop{ActorType: global.ACTOR_ROLE, ActorID: roleID, X: mapInfo.Pos.X, Y: mapInfo.Pos.Y, Dir: dir}
		state.SendRoleProto(roleID, proto)
	}
}

// 怪物停止移动要简单很多
var MoveMonsterStop = func(state *MapState, monsterID int64, notice bool) {
	mapInfo := state.MapMonsters[monsterID]
	if mapInfo == nil {
		return
	}
	aKey := AKey{ActorType: global.ACTOR_MONSTER, ActorID: monsterID}
	actor := state.GetActor(aKey)
	if actor.IsMove {
		actor.IsMove = false
		DelMove(state.MoveActor, aKey)
		mapInfo.MovePath = nil
		if notice {
			proto := &global.MapTocMoveStop{ActorType: global.ACTOR_MONSTER, ActorID: monsterID,
				X: mapInfo.Pos.X, Y: mapInfo.Pos.Y, Dir: mapInfo.Pos.Dir}
			state.BroadCastPos(mapInfo.Pos, proto)
		}
	}
}

// 每帧更新位置
func MoveUpdate(state *MapState, now2 int64) {
	for aKey, move := range state.MoveActor {
		actor := state.GetActor(aKey)
		if !ActorMove(state, actor, move, now2, true) {
			DelMove(state.MoveActor, aKey)
		}
	}
}

var ActorMove = func(state *MapState, actor *MapActor, move *Move, now2 int64, syncStop bool) (keepMove bool) {
	if !actor.IsMove {
		return false
	}
	dcTime := now2 - move.LastUpdateTime
	if dcTime < 0 {
		return true
	}
	mapInfo := state.GetMapInfo(actor.ActorType, actor.ActorID)
	moveSpeed, movePath, pos := mapInfo.GetMove()
	moveDistance := float64(moveSpeed) * float64(dcTime) / (100 * 1000)
	newMoveDistance := move.StepCount + moveDistance
	var x, y float32
	var dir int16
	var nexGrid bool
	if newMoveDistance >= move.StepTotal {
		// 走完当前拐点，判断一下是否有下一个拐点
		size := len(movePath.GridPath)
		switch size {
		case 0:
			// 走完了
			x = move.DestX
			y = move.DestY
			dir = pos.Dir
			move.LastUpdateTime = 0
			movePath = nil
		case 1:
			nexGrid = true
			startGrid := movePath.GridPath[0]
			// 最后一个拐点
			move.StepCount = newMoveDistance - move.StepTotal
			move.DestX = movePath.EndX
			move.DestY = movePath.EndY
			move.StepTotal = CalcTotalMove(float32(startGrid.GridX), float32(startGrid.GridY), movePath.EndX, movePath.EndY)
			x, y, dir = CalcNewXY(float32(startGrid.GridX), float32(startGrid.GridY), movePath.EndX, movePath.EndY, move.StepCount)
		default:
			nexGrid = true
			startGrid := movePath.GridPath[0]
			endGrid := movePath.GridPath[1]
			move.StepCount = newMoveDistance - move.StepTotal
			move.DestX = float32(endGrid.GridX)
			move.DestY = float32(endGrid.GridY)
			move.StepTotal = CalcTotalMove(float32(startGrid.GridX), float32(startGrid.GridY), move.DestX, move.DestY)
			x, y, dir = CalcNewXY(float32(startGrid.GridX), float32(startGrid.GridY), move.DestX, move.DestY, move.StepCount)
		}
	} else {
		move.StepCount = newMoveDistance
		x, y, dir = CalcNewXY(pos.X, pos.Y, movePath.EndX, movePath.EndY, moveDistance)
	}
	// 检查一下新的坐标点是否可走
	if !state.XYWalkAble(x, y) {
		// 客户端的路径可能会偶尔走过不可走的点，这里做一下容错
		if actor.ActorType == global.ACTOR_ROLE && actor.ActorID > 0 && moveDistance >= 300 {
			kernel.ErrorLog("un walkable:%d,%f,%f", state.Config.MapID, x, y)
			if syncStop {
				SyncStop(state, actor, false)
			}
			return false
		}
		return true
	}
	oldX := pos.X
	oldY := pos.Y
	pos.X = x
	pos.Y = y
	pos.Dir = dir
	if move.LastUpdateTime == 0 {
		mapInfo.SetMovePath(nil)
		actor.IsMove = false
		AoiUpdatePos(state, oldX, oldY, pos, mapInfo)
		return false
	}
	move.LastUpdateTime = now2
	if nexGrid {
		movePath.GridPath = movePath.GridPath[1:]
	}
	AoiUpdatePos(state, oldX, oldY, pos, mapInfo)
	return true
}

func CalcTotalMove(x, y, x2, y2 float32) float64 {
	x = x2 - x
	y = y2 - y
	return math.Sqrt(float64(x*x + y*y))
}

func CalcNewXY(x, y, x2, y2 float32, dist float64) (float32, float32, int16) {
	r := math.Atan2(float64(y2-y), float64(x2-x))
	x2 = float32(math.Cos(r)*dist) + x
	y2 = float32(math.Sin(r)*dist) + y
	dir := int16(math.Round(r * 180 / math.Pi))
	if dir < 0 {
		dir += 360
	}
	return x2, y2, dir
}

func SyncStop(state *MapState, actor *MapActor, del bool) {
	mapInfo := state.GetMapInfo(actor.ActorType, actor.ActorID)
	actor.IsMove = false
	if del {
		DelMove(state.MoveActor, AKey{ActorID: actor.ActorID, ActorType: actor.ActorType})
	}
	mapInfo.SetMovePath(nil)
	pos := mapInfo.GetPos()
	proto := &global.MapTocStop{ActorType: actor.ActorType, ActorID: actor.ActorID, X: pos.X, Y: pos.Y, Dir: pos.Dir}
	state.BroadCastPos(pos, proto)
}

func MoveCheckDist(state *MapState, actorType int8, actorID int64, pos *global.PPos, startX, startY float32) bool {
	startX = startX - pos.X
	startY = startY - pos.Y
	dist := startX*startX + startY*startY
	if dist <= DELTA_ERROR {
		if dist > DIST_FAIL && actorType == global.ACTOR_ROLE {
			state.Roles[actorID].MoveFail++
		}
		return true
	}
	return false
}

func DelMove(moveMap map[AKey]*Move, aKey AKey) {
	if move, ok := moveMap[aKey]; ok {
		_movePool.Put(move)
		delete(moveMap, aKey)
	}
}
