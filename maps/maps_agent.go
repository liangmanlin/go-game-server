package maps

import (
	"game/global"
	"game/lib"
	"github.com/liangmanlin/gootp/kernel"
	"unsafe"
)

var agentSvr = &kernel.Actor{
	Init: func(ctx *kernel.Context, pid *kernel.Pid, args ...interface{}) unsafe.Pointer {
		return nil
	},
	HandleCast: func(ctx *kernel.Context, msg interface{}) {
		switch m := msg.(type) {
		case *AgentChangeMap:
			if pid := lib.GetMapPid(m.MapName); pid != nil {
				lib.CastMap(pid,global.MAP_MOD_ROLE,m.Change)
			}else{
				// TODO 回城，这里比较复杂，先简单考虑直接连通
				lib.CastPlayer(m.Change.RoleData.Player,global.PLAYER_MOD_MAP,m.Change)
			}
		}

	},
	HandleCall: func(ctx *kernel.Context, request interface{}) interface{} {
		return nil
	},
	Terminate: func(ctx *kernel.Context, reason *kernel.Terminate) {

	},
	ErrorHandler: func(ctx *kernel.Context, err interface{}) bool {
		return true
	},
}
