package maps

import (
	"game/global"
	"github.com/liangmanlin/gootp/kernel"
)

func (m *MapState)BroadCastPos(pos *global.PPos,proto interface{})  {
	m.BroadCastAreas(Get9AreaByPos(pos), proto)
}

func (m *MapState)BroadCastXY(x,y float32,proto interface{})  {
	m.BroadCastAreas(Get9AreaByArea(GetAreaByXY(x,y)), proto)
}

func (m *MapState)BroadCastAreas(areas []Area, proto interface{}) {
	bin := Encoder.Encode(proto, 2)
	roleAreaMap := m.AreaMap[global.ACTOR_ROLE]
	for i := range areas {
		l := roleAreaMap.AreaActorIDs(areas[i])
		for _, roleID := range l {
			if r, ok := m.Roles[roleID]; ok {
				kernel.Cast(r.TcpPid, bin)
			}
		}
	}
}

func (m *MapState)BroadCastMapAreas(areas map[Area]bool, proto interface{}) {
	bin := Encoder.Encode(proto, 2)
	rm := m.AreaMap[global.ACTOR_ROLE]
	for area,_ := range areas {
		l := rm.AreaActorIDs(area)
		for _, roleID := range l {
			if r, ok := m.Roles[roleID]; ok {
				kernel.Cast(r.TcpPid, bin)
			}
		}
	}
}

func (m *MapState)BroadCastPosExclude(pos *global.PPos,roleID int64,proto interface{})  {
	m.BroadCastAreasExclude(Get9AreaByPos(pos),roleID, proto)
}

func (m *MapState)BroadCastAreasExclude(areas []Area,roleID int64, proto interface{}) {
	bin := Encoder.Encode(proto, 2)
	rm := m.AreaMap[global.ACTOR_ROLE]
	for i := range areas {
		l := rm.AreaActorIDs(areas[i])
		for _, rid := range l {
			if rid == roleID {
				continue
			}
			if r, ok := m.Roles[rid]; ok {
				kernel.Cast(r.TcpPid, bin)
			}
		}
	}
}

func (m *MapState)BroadCastRoles(roles []int64, proto interface{}) {
	bin := Encoder.Encode(proto, 2)
	for _, roleID := range roles {
		if r, ok := m.Roles[roleID]; ok {
			kernel.Cast(r.TcpPid, bin)
		}
	}
}

func (m *MapState) SendRoleProto(roleID int64, proto interface{}) {
	if r, ok := m.Roles[roleID]; ok {
		bin := Encoder.Encode(proto, 2)
		kernel.Cast(r.TcpPid, bin)
	}
}
