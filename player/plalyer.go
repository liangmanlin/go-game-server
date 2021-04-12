package player

import (
	"game/global"
	"game/proto"
	"github.com/liangmanlin/gootp/gate"
	"github.com/liangmanlin/gootp/gate/pb"
	"github.com/liangmanlin/gootp/kernel"
	"github.com/liangmanlin/gootp/timer"
	"unsafe"
)

var DEC *pb.Coder
var ENC *pb.Coder
var Router map[int]*global.HandleFunc
var modList = []mod{
	{bagName, &BagLoad, &BagPersistent},
	{attrName, &AttrLoad, &AttrPersistent},
}

const PersistentTime int64 = 10 * 60 * 1000 // 毫秒

var Actor = &kernel.Actor{
	Init: func(ctx *kernel.Context, pid *kernel.Pid, args ...interface{}) unsafe.Pointer {
		conn := args[0].(*gate.Conn)
		gwPid := args[1].(*kernel.Pid)
		roleID := args[2].(int64)
		player := global.Player{
			Conn:           conn,
			GWPid:          gwPid,
			RoleID:         roleID,
			DirtyMod:       make(map[string]bool),
			Timer:          timer.NewTimer(),
			PersistentTime: kernel.Now2() + PersistentTime,
		}
		player.BackupMap = make(map[global.BackupKey]int)
		// 加载数据
		for _, m := range modList {
			(*m.load)(ctx, &player)
		}
		kernel.ErrorLog("player start: %d",roleID)
		// 角色进程接收网络数据，减少一次消息转发
		conn.StartReaderDecode(pid, DEC.Decode)
		kernel.TimerStart(kernel.TimerTypeForever, pid, 1000, global.Loop{})
		return unsafe.Pointer(&player)
	},
	HandleCast: func(ctx *kernel.Context, msg interface{}) {
		switch m := msg.(type) {
		case gate.Pack:
			if !proto.Router(Router, m.ProtoID, m.Proto, ctx, (*global.Player)(ctx.State)) {
				kernel.ErrorLog("un handle id:%d msg: %#v", m.ProtoID, m.Proto)
			}
		case *gate.TcpError:
			kernel.ErrorLog("tcp error:%s", m.Err.Error())
			kernel.Cast((*global.Player)(ctx.State).GWPid, m)
		case global.Loop:
			PlayerLoop(ctx, (*global.Player)(ctx.State))
		}
	},
	HandleCall: func(ctx *kernel.Context, request interface{}) interface{} {
		return nil
	},
	Terminate: func(ctx *kernel.Context, reason *kernel.Terminate) {
		kernel.Cast((*global.Player)(ctx.State).GWPid, 1)
	},
	ErrorHandler: func(ctx *kernel.Context, err interface{}) bool {
		return true
	},
}

var PlayerLoop = func(ctx *kernel.Context, player *global.Player) {
	now2 := kernel.Now2()
	player.Timer.Loop(player, now2)
	// 判断是否需要持久化
	if now2 >= player.PersistentTime {
		player.PersistentTime += PersistentTime
		for _, m := range modList {
			if _,ok := player.DirtyMod[m.name];ok {
				(*m.persistent)(ctx, player)
				delete(player.DirtyMod,m.name)
			}
		}
	}
}

// 非常简单的包装，不可能有热更需求
func StartTimer(player *global.Player, key, id, inv, times int32, f interface{}, arg ...interface{}) {
	player.Timer.Add(timer.TimerKey{Key: key, ID: id}, inv, times, f, arg...)
}
