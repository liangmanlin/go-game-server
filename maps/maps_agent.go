package maps

import (
	"game/global"
	"game/lib"
	"github.com/liangmanlin/gootp/kernel"
)

var agentSvr = kernel.NewActor(
	kernel.HandleCastFunc(
		func(ctx *kernel.Context, msg interface{}) {
			switch m := msg.(type) {
			case *AgentChangeMap:
				if pid := lib.GetMapPid(m.MapName); pid != nil {
					lib.CastMap(pid, global.MAP_MOD_ROLE, m.Change)
				} else {
					// TODO 回城，这里比较复杂，先简单考虑直接连通
					lib.CastPlayer(m.Change.RoleData.Player, global.PLAYER_MOD_MAP, m.Change)
				}
			}

		}),
)
