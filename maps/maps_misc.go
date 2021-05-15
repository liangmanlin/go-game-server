package maps

import (
	"game/global"
	"game/lib"
)

func MakeRandomXY(state *MapState, pos *global.PPos, dist int32) (x, y float32, ok bool) {
	// 最多随机20次
	for i := 0; i < 20; i++ {
		rx := state.Rand.Random(-dist, dist)
		ry := state.Rand.Random(-dist, dist)
		if rx != 0 && ry != 0 {
			x = pos.X + float32(rx)
			y = pos.Y + float32(ry)
			if state.XYWalkAble(x, y) {
				ok = true
				break
			}
		}
	}
	return
}

func CalcNewWalkAbleXY(state *MapState, pos *global.PPos, dir int16, dist int32) (x, y float32) {
	cos := lib.Cos(dir)
	sin := lib.Sin(dir)
	x = pos.X + float32(cos*float64(dist))
	y = pos.Y + float32(sin*float64(dist))
	// 假如目标点可走，直接返回
	if state.XYWalkAble(x, y) {
		return
	}
	for i := int32(1); i < dist; i++ {
		x = pos.X + float32(cos*float64(i))
		y = pos.Y + float32(sin*float64(i))
		if state.XYWalkAble(x, y) {
			return
		}
	}
	return pos.X, pos.Y
}
