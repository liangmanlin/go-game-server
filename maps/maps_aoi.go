package maps

import (
	"game/global"
	"github.com/liangmanlin/gootp/gutil"
	"sync"
	"unsafe"
)

/*
使用九宫格aoi
*/

var _mapInfoPool = sync.Pool{
	New: func()interface {}{
		return make([]MapInfo,0,5)
	},
}

var areaCache = make(map[Area][]Area, 100)

// 根据九宫格特性，对于同一个变编号的slice，他周围的8个slice id是一样的，所以为了节约内存，这里直接cache
func buildAreasCache(maxWidth, maxHeight int32) {
	maxWidthID := maxWidth / AREA_WIDTH
	maxHeightID := maxHeight / AREA_HEIGHT
	for i := int32(0); i <= maxWidthID; i++ {
		for j := int32(0); j <= maxHeightID; j++ {
			areaID := GetArea(i, j)
			areas := getAroundAreas(i, j, maxWidthID, maxHeightID)
			areaCache[areaID] = areas
		}
	}
}

// 返回一个完整的九宫格
func getAroundAreas(areaX, areaY, maxWidthID, maxHeightID int32) []Area {
	xMin := gutil.MaxInt32(areaX-1, 0)
	xMax := gutil.MinInt32(areaX+1, maxWidthID)
	yMin := gutil.MaxInt32(areaY-1, 0)
	yMax := gutil.MinInt32(areaY+1, maxHeightID)
	areas := make([]Area, (xMax-xMin+1)*(yMax-yMin+1))
	var idx int
	for i := xMin; i <= xMax; i++ {
		for j := yMin; j <= yMax; j++ {
			areas[idx] = GetArea(i, j)
			idx++
		}
	}
	return areas
}

func Get9AreaByPos(pos *global.PPos) []Area {
	area := GetAreaByPos(pos)
	return areaCache[area]
}

func Get9AreaByArea(area Area) []Area {
	return areaCache[area]
}

func GetAreaByPos(pos *global.PPos) Area {
	return GetArea(gutil.Round(pos.X), gutil.Round(pos.Y))
}

func GetAreaByXY(posX, posY float32) Area {
	return GetArea(gutil.Round(posX), gutil.Round(posY))
}

func GetAreaByGrid(grid Grid) Area {
	return GetArea(grid.X, grid.Y)
}

func GetArea(gridX, gridY int32) Area {
	return Area{X: int16(gridX / AREA_WIDTH), Y: int16(gridY / AREA_HEIGHT)}
}

func GetGridByPos(pos *global.PPos) Grid {
	return Grid{X: gutil.Round(pos.X), Y: gutil.Round(pos.Y)}
}

func GetGrid(posX, posY float32) Grid {
	return Grid{X: gutil.Round(posX), Y: gutil.Round(posY)}
}

func EnterArea(state *MapState, area Area, actorType int8, actorID int64) {
	state.AreaMap[actorType].EnterActorID(area, actorID)
}

func LeaveArea(state *MapState, area Area, actorType int8, actorID int64) {
	state.AreaMap[actorType].LeaveActorID(area, actorID)
}

type AreaActors []int64

func NewAreaActors() *AreaActors {
	a := make(AreaActors, 0, 2)
	return &a
}

func (a *AreaActors) Del(actorID int64) {
	tmp := *a
	size := len(tmp) - 1
	for i := 0; i <= size; i++ {
		if tmp[i] == actorID {
			tmp[i] = tmp[size]
			*a = tmp[0:size] // 暂时不考虑gc问题，理论上不会占多少内存
			break
		}
	}
}
func (a *AreaActors) Add(actorID int64) {
	*a = append(*a, actorID)
}

type AreaMap map[Area]*AreaActors

func (a AreaMap) EnterActorID(area Area, actorID int64) {
	if p, ok := a[area]; ok {
		p.Add(actorID)
	}
	p := NewAreaActors()
	p.Add(actorID)
	a[area] = p
}

func (a AreaMap) LeaveActorID(area Area, actorID int64) {
	if p, ok := a[area]; ok {
		p.Del(actorID)
	}
}

func (a AreaMap) AreaListActorIDs(areas []Area) []int64 {
	// 这里假设一个基数，大概4个人
	rs := make([]int64, 0, 4)
	for i := range areas {
		if p, ok := a[areas[i]]; ok {
			rs = append(rs, *p...)
		}
	}
	return rs
}

// 只读，上层逻辑需要注意,而且不建议存储，每次使用都读取
func (a AreaMap) AreaActorIDs(area Area) []int64 {
	if p, ok := a[area]; ok {
		return *p
	}
	return nil
}

func AoiUpdatePos(state *MapState, oldX, oldY float32, newPos *global.PPos, mapInfo MapInfo) {
	oldGrid := GetGrid(oldX, oldY)
	newGrid := GetGridByPos(newPos)
	oldArea := GetAreaByGrid(oldGrid)
	newArea := GetAreaByGrid(newGrid)
	actorType := mapInfo.Type()
	actorID := mapInfo.ID()
	// 只有改变了区域才需要更新视野目标
	if oldArea != newArea {
		// 先退出
		state.AreaMap[actorType].LeaveActorID(oldArea, actorID)
		leaveAreas, enterAreas := GetEnterLeave(oldArea, newArea) // 计算一个进入和离开的区域
		LeaveRoles := GetAreasActorIDList(state, leaveAreas, global.ACTOR_ROLE)
		// 先通知退出
		state.BroadCastRoles(LeaveRoles, &global.MapTocActorLeaveArea{ActorID: actorID, ActorType: actorType})
		//角色需要看见新的视野，但是怪物是不需要的
		if actorType == global.ACTOR_ROLE {
			roles := GetAreaRoles(state, enterAreas)
			monsters := GetAreaMonsters(state, enterAreas)
			LeaveMonsters := GetAreasActorIDList(state, leaveAreas, global.ACTOR_MONSTER)
			proto := &global.MapTocEnterArea{RolesShow: roles,
				MonstersShow: monsters,
				RolesDel:     LeaveRoles,
				MonstersDel:  LeaveMonsters,
			}
			state.SendRoleProto(actorID, proto)
			proto2 := &global.MapTocRoleEnterArea{MapInfo: mapInfo.(*global.PMapRole)}
			state.BroadCastAreas(enterAreas, proto2)
		} else {
			proto2 := &global.MapTocMonsterEnterArea{MapInfo: mapInfo.(*global.PMapMonster)}
			state.BroadCastAreas(enterAreas, proto2)
		}
		// 最后进入area
		state.AreaMap[actorType].EnterActorID(newArea, actorID)
	}
	if actorType == global.ACTOR_MONSTER && oldGrid != newGrid {
		// 切换一下计数器
		state.PosMonsterUP(oldGrid, -1)
		state.PosMonsterUP(newGrid, 1)
	}
}

// 通过计算得到
func GetEnterLeave(leaveArea, enterArea Area) (leave []Area, enter []Area) {
	key := Vector(leaveArea, enterArea)
	if p, ok := constLeaveEnterMap[key]; ok {
		size := len(p.leave)
		leave = make([]Area, 0, size)
		s := p.leave
		for i := range s {
			t := s[i]
			x := t.X + leaveArea.X
			y := t.Y + leaveArea.Y
			if x >= 0 && y >= 0 {
				leave = append(leave, Area{x, y})
			}
		}
		size = len(p.enter)
		enter = make([]Area, 0, size)
		s = p.enter
		for i := range s {
			t := s[i]
			x := t.X + enterArea.X
			y := t.Y + enterArea.Y
			if x >= 0 && y >= 0 {
				enter = append(enter, Area{x, y})
			}
		}
		return leave, enter
	}
	// 到这里说明跨越了九宫格,直接读取预生成的信息
	// TODO 看一下后续设计，理论上这里不需要拷贝
	copy(leave, areaCache[leaveArea])
	copy(enter, areaCache[enterArea])
	return leave, enter
}

func Vector(old, new Area) Area {
	return Area{new.X - old.X, new.Y - old.Y}
}

func GetAreasActorIDList(state *MapState, areas []Area, actorType int8) []int64 {
	return state.AreaMap[actorType].AreaListActorIDs(areas)
}

// 通常这些返回的对象都是不能修改的，切记
func GetAreaRoles(state *MapState, areas []Area) []*global.PMapRole {
	p := state.AreaMap[global.ACTOR_ROLE]
	rs := make([]*global.PMapRole, 0, 2)
	mr := state.MapRoles
	for i := range areas {
		if l, ok := p[areas[i]]; ok {
			for _, roleID := range *l {
				if mapInfo, ok := mr[roleID]; ok {
					rs = append(rs, mapInfo)
				}
			}
		}
	}
	return rs
}

func GetAreaMonsters(state *MapState, areas []Area) []*global.PMapMonster {
	p := state.AreaMap[global.ACTOR_MONSTER]
	rs := make([]*global.PMapMonster, 0, 2)
	mr := state.MapMonsters
	for i := range areas {
		if l, ok := p[areas[i]]; ok {
			for _, monsterID := range *l {
				if mapInfo, ok := mr[monsterID]; ok {
					rs = append(rs, mapInfo)
				}
			}
		}
	}
	return rs
}

// 返回值最好不要进行修改，并且使用完归还pool里面
func AreasActorFold(state *MapState, actorType int8, areas []Area,srcInfo MapInfo, f AreaFoldFunc, args ...unsafe.Pointer) []MapInfo {
	rs := MakeMapInfoSlice()
	pm := state.AreaMap[actorType]
	for i := range areas {
		if l, ok := pm[areas[i]]; ok {
			for _, actorID := range *l {
				if mapInfo := state.GetMapInfo(actorType, actorID); mapInfo != nil {
					if f(state,srcInfo,mapInfo,args...) {
						rs = append(rs, mapInfo)
					}
				}
			}
		}
	}
	return rs
}

func MakeMapInfoSlice() []MapInfo {
	s := _mapInfoPool.Get().([]MapInfo)
	return s[0:0]
}

func ReleaseMapInfoSlice(s []MapInfo)  {
	_mapInfoPool.Put(s)
}
