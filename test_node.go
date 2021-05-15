package main

import (
	"game/global"
	_ "game/player"
	"github.com/liangmanlin/gootp/kernel"
	"github.com/liangmanlin/gootp/node"
)

func main() {
	kernel.Env.WriteLogStd = true
	kernel.Env.LogPath = "./logs"
	kernel.Env.SetTimerMinTick(50)
	// 设置Env需要在启动之前
	kernel.KernelStart(func() {
		node.Env.Port = 5000 // 可以指定端口
		node.Env.PingTick = 30000 // 指定ping频率 毫秒
		node.Start("game_2@127.0.0.1", "6d27544c07937e4a7fab8123291cc4df",
			[]interface{}{&global.TestCall{},&global.StrArgs{},&global.RpcStrResult{}})
		destNode := "game@127.0.0.1"
		if node.ConnectNode(destNode) {
			succ,rss := node.RpcCall("game@127.0.0.1","Echo",&global.StrArgs{Str: "ffffffffffffffff"})
			kernel.ErrorLog("%v,%#v",succ,rss)
			kernel.CastNameNode("ttt",destNode,&global.TestCall{ID: 1})
			ok,rs := kernel.CallNameNode("ttt",destNode,&global.TestCall{ID: 2})
			kernel.ErrorLog("%v,%#v",ok,rs)
			for _,n := range kernel.Nodes(){
				kernel.ErrorLog("%#v",n)
			}
		}
	},
	nil)
}


