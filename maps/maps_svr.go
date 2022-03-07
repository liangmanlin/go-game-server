package maps

import (
	"game/lib"
	"github.com/liangmanlin/gootp/kernel"
)

func StartMap(mapID int32, mapName string, mod *MapMod,args...interface{}) {
	kernel.SupStartChild("map_sup", &kernel.SupChild{
		ChildType: kernel.SupChildTypeWorker,
		ReStart:   true,
		Svr:       SvrActor,
		InitArgs:  append(kernel.MakeArgs(mapID, mapName, mod),args...),
	})
}

func StartCopy(mapID int32, mapName string, mod *MapMod,args...interface{})  {
	kernel.Start(SvrActor,append(kernel.MakeArgs(mapID, mapName, mod),args...)...)
}

var SvrActor = &kernel.Actor{
	Init: func(ctx *kernel.Context, pid *kernel.Pid, args ...interface{}) interface{} {
		mapID := args[0].(int32)
		mapName := args[1].(string)
		mod := args[2].(*MapMod)
		state := NewMapState(mapID,mapName,mod,ctx)
		if len(args) > 3 {
			mod.Init(state,ctx, mapID, args[3:]...)
		}else{
			mod.Init(state,ctx, mapID)
		}
		lib.RegisterMap(mapName,pid)
		// 启动定轮训
		kernel.SendAfterForever(pid,100,kernel.Loop{})
		return state
	},
	HandleCast: func(ctx *kernel.Context, msg interface{}) {
		state := State(ctx)
		switch msg.(type) {
		case kernel.Loop:
			mapLoop(state,ctx)
		}
	},
	HandleCall: func(ctx *kernel.Context, request interface{}) interface{} {
		return nil
	},
	Terminate: func(ctx *kernel.Context, reason *kernel.Terminate) {
		lib.UnRegisterMap(State(ctx).Name)
	},
	ErrorHandler: func(ctx *kernel.Context, err interface{}) bool {
		return true
	},
}

func mapLoop(state *MapState,ctx *kernel.Context)  {
	now2 := kernel.Now2()
	MoveUpdate(state,now2)
	SkillUpdate(state,now2)
	if (state.LC & 1) == 0{ // 200ms 一次
		BuffLoop(state,now2)
		MonsterLoop(state,now2)
		state.Timer.Loop(state,now2) // 驱动计时器
	}else {
		if state.LC == 9 {
			// 秒级轮询
			state.LC = -1
			RoleSecondLoop(state, now2)
			MonsterUpdateSecond(state,now2)
		}
	}
	state.LC++
}