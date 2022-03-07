package player

import (
	"game/global"
	"game/lib"
	"game/proto"
	"github.com/liangmanlin/gootp/bpool"
	"github.com/liangmanlin/gootp/gate"
	"github.com/liangmanlin/gootp/gate/pb"
	"github.com/liangmanlin/gootp/kernel"
	"github.com/liangmanlin/gootp/timer"
	"unsafe"
)

var modRouter [global.PLAYER_MOD_MAX]*MsgHandler
var DEC *pb.Coder
var ENC *pb.Coder
var Router map[int]*global.HandleFunc
var modList = []mod{
	{bagName, &BagLoad, &BagPersistent},
	{attrName, &AttrLoad, &AttrPersistent},
	{propName, &PropLoad, &PropPersistent},
	{mapModName, &MapLoad, &MapPersistent},
}

const PersistentTime int64 = 10 * 60 * 1000 // 毫秒

var Actor = &kernel.Actor{
	Init: func(ctx *kernel.Context, pid *kernel.Pid, args ...interface{}) interface{} {
		conn := args[0].(gate.Conn)
		gwPid := args[1].(*kernel.Pid)
		roleID := args[2].(int64)
		player := global.Player{
			Conn:           conn,
			GWPid:          gwPid,
			RoleID:         roleID,
			DirtyMod:       make(map[string]bool),
			Timer:          timer.NewTimer(),
			PersistentTime: kernel.Now2() + PersistentTime,
			PropData:       unsafe.Pointer(lib.NewPropData()),
		}
		player.Context = ctx
		player.BackupMap = make(map[global.BackupKey]int)
		// 加载数据
		for _, m := range modList {
			(*m.load)(&player)
		}
		// 计算属性
		InitProps(&player)
		kernel.ErrorLog("Player start: %d", roleID)
		// 角色进程接收网络数据，减少一次消息转发
		conn.StartReader(pid)
		kernel.SendAfterForever(pid, 1000, kernel.Loop{})
		lib.SetRolePid(roleID, pid)

		// 最后进入地图
		MapFirstEnter(&player)
		return &player
	},
	HandleCast: func(ctx *kernel.Context, msg interface{}) {
		switch m := msg.(type) {
		case gate.Pack:
			if !proto.Router(Router, m.ProtoID, m.Proto, Player(ctx)) {
				kernel.ErrorLog("un handle id:%d msg: %#v", m.ProtoID, m.Proto)
			}
		case *kernel.KMsg:
			if int(m.ModID) < len(modRouter) {
				(*modRouter[m.ModID])(Player(ctx), m.Msg)
			} else {
				kernel.UnHandleMsg(m)
			}
		case *gate.TcpError:
			kernel.ErrorLog("tcp error:%s", m.Err.Error())
			kernel.Cast(Player(ctx).GWPid, m)
		case kernel.Loop:
			loop(Player(ctx))
		case global.TcpReConnect:
			// TODO 刷新一下心跳
			Player(ctx).HeartTime = kernel.Now2()
		case *bpool.Buff:
			protoID,p := DEC.Decode(m.ToBytes())
			if !proto.Router(Router, protoID, p, Player(ctx)) {
				kernel.ErrorLog("un handle id:%d msg: %#v", protoID,p)
			}
			m.Free()
		default:
			kernel.UnHandleMsg(msg)
		}
	},
	HandleCall: func(ctx *kernel.Context, request interface{}) interface{} {
		player := Player(ctx)
		switch request.(type) {
		case global.Kick:
			// 先退出场景
			MapExit(player)
			//可以优先退出网络
			kernel.Cast(player.GWPid, 1)
		}
		return nil
	},
	Terminate: func(ctx *kernel.Context, reason *kernel.Terminate) {
		player := Player(ctx)
		// 先执行退出流程
		HookExit(player)
		// 执行持久化
		player.PersistentTime = 0
		persistent(player, kernel.Now2())
		lib.DelRolePid(player.RoleID)
		kernel.Cast(player.GWPid, 1)
	},
	ErrorHandler: func(ctx *kernel.Context, err interface{}) bool {
		return true
	},
}

func loop(player *global.Player) {
	now2 := kernel.Now2()
	player.Timer.Loop(player, now2)
	// 判断是否需要持久化
	persistent(player, now2)
}

func persistent(player *global.Player, now2 int64) {
	if now2 >= player.PersistentTime {
		player.PersistentTime += PersistentTime
		for _, m := range modList {
			if _, ok := player.DirtyMod[m.name]; ok {
				(*m.persistent)(player)
				delete(player.DirtyMod, m.name)
			}
		}
	}
}

// 非常简单的包装，不可能有热更需求
func StartTimer(player *global.Player, key, id, inv, times int32, f interface{}, arg ...interface{}) {
	player.Timer.Add(timer.TimerKey{Key: key, ID: id}, inv, times, f, arg...)
}

func DelTimer(player *global.Player, key, id int32) {
	player.Timer.Del(timer.TimerKey{Key: key, ID: id})
}

func Player(ctx *kernel.Context) *global.Player {
	return ctx.State.(*global.Player)
}

func Start(gwPid *kernel.Pid, conn gate.Conn, roleID int64) *kernel.Pid {
	err, pid := kernel.SupStartChild("player_sup",
		&kernel.SupChild{
			ChildType: kernel.SupChildTypeWorker,
			ReStart:   false,
			Svr:       Actor,
			Name:      lib.GetPlayerName(roleID),
			InitArgs:  kernel.MakeArgs(conn, gwPid, roleID),
		})
	if err != nil {
		kernel.ErrorLog("start Player error:%#v", err)
		return nil
	}
	return pid
}

func SendProto(player *global.Player, proto interface{}) {
	bin := ENC.Encode(proto, 2)
	kernel.Cast(player.GWPid, bin)
}
