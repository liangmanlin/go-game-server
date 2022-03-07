package main

import (
	"fmt"
	"game/boot"
	"game/config"
	"game/global"
	"game/gw"
	"game/lib"
	"game/maps"
	_ "game/maps/mod"
	"game/player"
	"game/proto"
	"game/proto/router"
	"game/service"
	"github.com/liangmanlin/gootp/db"
	"github.com/liangmanlin/gootp/gate"
	"github.com/liangmanlin/gootp/gate/pb"
	"github.com/liangmanlin/gootp/kernel"
	"github.com/liangmanlin/gootp/node"
	"plugin"
	"time"
)

func main() {
	root := lib.GetServerRoot()
	config.LoadAll(root + "/config/json")
	//kernel.Env.SetTimerMinTick(100)
	// 设置Env需要在启动之前
	kernel.KernelStart(initGame, stopGame)
}

func initGame() {
	kernel.ErrorLog("game is debug: %v", global.DEBUG)
	node.Env.PingTick = 30 * 1000 // 可以指定心跳间隔,默认60*1000ms
	global.NodeProto = append(global.NodeProto, maps.NodeProto()...)
	node.StartFromCommandLine(global.NodeProto)
	global.NodeProto = nil // gc
	kernel.SupStartChild("kernel", &kernel.SupChild{ChildType: kernel.SupChildTypeSup, Name: "game_sup"})
	_, serviceSup := kernel.SupStartChild("game_sup", &kernel.SupChild{ChildType: kernel.SupChildTypeSup, Name: "service_sup"})
	player.ENC = pb.Parse(proto.TOC, 0)
	player.DEC = pb.Parse(proto.TOS, 1)
	// 删除没有的数据，以免浪费gc的时间
	proto.TOC = nil
	proto.TOS = nil
	player.Router = router.MakeRouter()
	dbConfig := config.Server.Get("db").([]interface{})
	port := int(dbConfig[1].(float64))
	DBConfig := db.Config{Host: dbConfig[0].(string), Port: port, User: dbConfig[2].(string), PWD: dbConfig[3].(string)}
	lib.GameDB = db.Start(0,DBConfig, global.DBTables, config.Server.GetStr("game_db"),5,db.MODE_NORMAL)
	maps.Start(player.ENC, lib.GetServerRoot()+"/config/maps")
	// 启动控制台
	kernel.StartConsole(
		kernel.ConsoleHandler("reload", reload,
			kernel.ConsoleArg("dir"),
			kernel.ConsoleCommit("热更新so")),
		kernel.ConsoleHandler("reload_config", reloadConfig,
			kernel.ConsoleArg("configName"),
			kernel.ConsoleCommit("热更新配置")),
	)
	svrList := []boot.Boot{
		{service.AccountActor, global.SYS_ACCOUNT_SERVER},
	}
	boot.StartService(serviceSup, svrList)
	// 最后启动网关
	gatePort := config.Server.GetInt("game_port")
	gate.Start("gw", gw.TcpClientActor, gatePort, gate.WithUseEpoll(),gate.WithClientArgs(player.ENC))
	kernel.ErrorLog("start completed")
}

func stopGame() {
	gate.Stop("gw")
	time.Sleep(time.Second)
	kernel.ErrorLog("begin to stop map")
	kernel.AppStop("MapApp")
}

func reload(echo func(s string), commands []string) string {
	mod := commands[0]
	file := kernel.GetMainRoot() + "/hotfix/" + mod + "/" + mod + ".so"
	p, e := plugin.Open(file)
	if e != nil {
		kernel.ErrorLog("%s", e.Error())
		return fmt.Sprintf("so can not load: %s\n", file)
	}
	f, err := p.Lookup("HotFix")
	if err != nil {
		kernel.ErrorLog("%s", err.Error())
		return "HotFix func can not load\n"
	}
	f.(func())()
	kernel.ErrorLog("reload: %s", file)
	return fmt.Sprintf("reload: %s\n", file)
}

func reloadConfig(echo func(s string), commands []string) string {
	name := commands[0]
	var ok bool
	if name == "Server" {
		ok = config.Load(name, "./config/json/system")
	} else {
		ok = config.Load(name, "./config/json")
	}
	if ok {
		kernel.ErrorLog("reload config: %s", name)
		return fmt.Sprintf("reload config: %s\n", name)
	} else {
		kernel.ErrorLog("error: cannot reload config: %s", name)
		return fmt.Sprintf("error: cannot reload config: %s\n", name)
	}
}
