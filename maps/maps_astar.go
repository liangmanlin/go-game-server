package maps

import (
	"game/global"
	"github.com/liangmanlin/gootp/astar"
)

func AStarSearch(state *MapState, x, y, dx, dy float32) (astar.ReturnType, *global.PMovePath) {
	rt, gridList := astar.Search(state, x, y, dx, dy)
	if rt == astar.ASTAR_FOUNDED {
		size := len(gridList)
		movePath := &global.PMovePath{EndX: dx, EndY: dy}
		if size > 0 {
			pGridList := make([]*global.PGrid, 0, size/2)
			for i := 0; i < size; i += 2 {
				pGrid := &global.PGrid{GridX: gridList[i],GridY: gridList[i+1]}
				pGridList = append(pGridList,pGrid)
			}
			movePath.GridPath = pGridList
		}
		return rt,movePath
	}
	return rt, nil
}
