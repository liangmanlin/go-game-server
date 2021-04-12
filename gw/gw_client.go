package gw

import (
	"game/global"
	"game/player"
	"github.com/liangmanlin/gootp/gate"
	"github.com/liangmanlin/gootp/gate/pb"
	"github.com/liangmanlin/gootp/kernel"
	"unsafe"
)

var TcpClientActor = &kernel.Actor{
	Init: func(ctx *kernel.Context,pid *kernel.Pid, args ...interface{}) unsafe.Pointer {
		t := global.TcpClientState{}
		t.Coder = args[0].(*pb.Coder)
		t.Conn = args[1].(*gate.Conn)
		kernel.ErrorLog("connect")
		return unsafe.Pointer(&t)
	},
	HandleCast: func(context *kernel.Context, msg interface{}) {
		t := (*global.TcpClientState)(context.State)
		switch m := msg.(type) {
		case bool:
			doLogin(t, context)
		case []byte:
			if t.Conn == nil {
				return
			}
			err := t.Conn.SendBufHead(m)
			if err != nil {
				t.Conn.Close()
				t.Conn = nil
			}
		case *gate.TcpError:
			context.Exit("normal")
		case int:
			context.Exit("normal")
		default:
			kernel.ErrorLog("un handle msg: %#v", m)
		}
	},
	HandleCall: func(context *kernel.Context, request interface{}) interface{} {
		return nil
	},
	Terminate: func(context *kernel.Context, reason *kernel.Terminate) {
		t := (*global.TcpClientState)(context.State)
		if t.Conn != nil {
			t.Conn.Close()
		}
		kernel.ErrorLog("exit reason:%s", reason.Reason)
	},
	ErrorHandler: func(context *kernel.Context, err interface{}) bool {
		return false
	},
}

func doLogin(t *global.TcpClientState, context *kernel.Context) {
	t.Conn.SetHead(2)
	//err,buf := t.Conn.Recv(0,5000)
	//if err != nil {
	//	kernel.ErrorLog(err.Error())
	//	t.Conn.Close()
	//	context.Exit("normal")
	//	return
	//}
	//id,proto := pb.GetCoder(1).Decode(buf)
	//kernel.ErrorLog("%d,%#v",id,proto)

	// 该进程只负责发送网络数据
	// 玩家进程负责收
	var roleID int64 = 1
	kernel.Start(player.Actor, t.Conn, context.Self(),roleID)
}
