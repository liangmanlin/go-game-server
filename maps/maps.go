package maps

import (
	"game/global"
	"game/lib"
	"github.com/liangmanlin/gootp/astar"
	"github.com/liangmanlin/gootp/gate/pb"
	"github.com/liangmanlin/gootp/gutil"
	"github.com/liangmanlin/gootp/kernel"
	"github.com/liangmanlin/gootp/rand"
	"github.com/liangmanlin/gootp/timer"
	"log"
)

var modRouter [global.MAP_MOD_MAX]*MsgHandler

type MsgHandler func(ctx *kernel.Context, state *MapState, msg interface{})

var Encoder *pb.Coder

var ModMap = make(map[string]*MapMod)

func Start(encoder *pb.Coder, mapConfigPath string) {
	Encoder = encoder
	if err := kernel.AppStart(&app{mapConfigPath: mapConfigPath}); err != nil {
		log.Panic(err)
	}
}

func NewMapState(mapID int32, name string, mod *MapMod) *MapState {
	state := MapState{Config: GetMapConfig(mapID), Mod: mod, Name: name,
		MapRoles:    make(map[int64]*global.PMapRole),
		MapMonsters: make(map[int64]*global.PMapMonster),
		Roles:       make(map[int64]*MapRoleData),
		Monsters:    make(map[int64]*MState),
		Actors:      make(map[AKey]*MapActor),
		MoveActor:   make(map[AKey]*Move),
		Rand:        rand.New(),
		Timer:       timer.NewTimer(),
		DataMonster: &DataMonster{
			PosMonsterNum: make(map[Area]int32, 50),
		},
		DataSkill: &DataSkill{
			EntityList: make(map[int32]*SkillEntity, 8),
			Add:        make(map[int32]*SkillEntity, 4),
		},
	}
	for i := 1; i < len(state.AreaMap); i++ {
		state.AreaMap[i] = make(AreaMap, 5)
	}
	return &state
}

func State(ctx *kernel.Context) *MapState {
	return (*MapState)(ctx.State)
}

func (m *MapState) EnterActor(area Area, actorType int8, actorID int64) {
	m.AreaMap[actorType].EnterActorID(area, actorID)
}

func (m *MapState) LeaveActor(area Area, actorType int8, actorID int64) {
	m.AreaMap[actorType].LeaveActorID(area, actorID)
}

func (m *MapState) GetActor(aKey AKey) *MapActor {
	return m.Actors[aKey]
}

func (m *MapState) GetActorByID(actorType int8, actorID int64) *MapActor {
	return m.Actors[AKey{ActorID: actorID, ActorType: actorType}]
}

func (m *MapState) GetMapInfo(actorType int8, actorID int64) MapInfo {
	if actorType == global.ACTOR_ROLE {
		return m.MapRoles[actorID]
	}
	return m.MapMonsters[actorID]
}

func (m *MapState) GetMapInfoAKey(key AKey) MapInfo {
	return m.GetMapInfo(key.ActorType, key.ActorID)
}

func (m *MapState) GetMapInfoRole(roleID int64) *global.PMapRole {
	return m.MapRoles[roleID]
}

func (m *MapState) GetMapInfoMonster(monsterID int64) *global.PMapMonster {
	return m.MapMonsters[monsterID]
}

func (m *MapState) PosWalkable(pos *global.PPos) bool {
	return m.Config.PosWalkAble(pos)
}

func (m *MapState) XYWalkAble(X, Y float32) bool {
	return m.Config.XYWalkAble(X, Y)
}

func (m *MapState) XYI32WalkAble(X, Y int32) bool {
	return m.Config.XYI32WalkAble(X, Y)
}

func (m *MapState) GetGridConfig() astar.GridConfig {
	return m.Config
}

func (m *MapState) GetAStarCache() *astar.AStar {
	return m.AStar
}

func (m *MapState) SetAStarCache(a *astar.AStar) {
	m.AStar = a
}

func (m *MapState) MakeMonsterID() int64 {
	dm := m.DataMonster
	dm.MonsterID++
	return dm.MonsterID
}

func (m *MapState) UpdateMonsterInfo(monsterID int64, pos *global.PPos, ul []*global.PKV) {
	proto := &global.MapTocUpdateMonsterInfo{ID: monsterID, UL: ul}
	m.BroadCastPos(pos, proto)
}

func (m *MapState) UpdateRoleInfo(roleID int64, pos *global.PPos, ul []*global.PKV) {
	proto := &global.MapTocUpdateRoleInfo{RoleID: roleID, UL: ul}
	m.BroadCastPos(pos, proto)
}

func (m *MapState) MonsterActor(monsterID int64) *MapActor {
	return m.Actors[AKey{monsterID, global.ACTOR_MONSTER}]
}

func (m *MapState) RoleActor(roleID int64) *MapActor {
	return m.Actors[AKey{roleID, global.ACTOR_ROLE}]
}

func (m *MapState) PosHaveMonster(pos *global.PPos) bool {
	dm := m.DataMonster
	k := Area{int16(gutil.Round(pos.X)), int16(gutil.Round(pos.Y))}
	if v, ok := dm.PosMonsterNum[k]; ok {
		return v > 0
	}
	return false
}

func (m *MapState) PosMonsterUP(grid Grid, inc int32) {
	mk := m.DataMonster.PosMonsterNum
	k := Area{int16(grid.X), int16(grid.Y)}
	if v, ok := mk[k]; ok {
		mk[k] = v + inc
	} else {
		mk[k] = gutil.MaxInt32(0, v+inc)
	}
}

func (m *MapState) PosSafe(pos *global.PPos) bool {
	return m.Config.PosSafe(pos)
}

func (m *MapState) AddSkillEntity(skillEntity *SkillEntity) {
	data := m.DataSkill
	if data.IsLoop {
		data.Add[skillEntity.ID] = skillEntity
	} else {
		data.EntityList[skillEntity.ID] = skillEntity
	}
}

func (m *MapState) DelSkillEntity(ID int32) {
	data := m.DataSkill
	if data.IsLoop {
		data.Del = append(data.Del, ID)
	} else {
		delete(data.EntityList, ID)
	}
}

func (m *MapState) CastPlayer(roleID int64, mod int32, msg interface{}) {
	if player := m.Roles[roleID]; player != nil {
		lib.CastPlayer(player.Player, mod, msg)
	}
}
