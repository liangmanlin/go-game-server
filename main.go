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
	"net/http"
	"plugin"
	"time"
)

func main() {
	root := lib.GetServerRoot()
	config.LoadAll(root + "/config/json")
	kernel.Env.WriteLogStd = true
	kernel.Env.LogPath = config.Server.GetStr("logs_dir")
	kernel.Env.SetTimerMinTick(100)
	// 设置Env需要在启动之前
	kernel.KernelStart(initGame, stopGame)
}

func initGame() {
	kernel.ErrorLog("game is debug: %v",global.DEBUG)
	node.Env.PingTick = 30 * 1000 // 可以指定心跳间隔,默认60*1000ms
	Cookie := config.Server.GetListStr("node", 1)
	global.NodeProto = append(global.NodeProto,maps.NodeProto()...)
	node.Start(config.Server.GetListStr("node", 0), Cookie, global.NodeProto)
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
	db.Start(DBConfig, global.DBTables, config.Server.GetStr("game_db"), config.Server.GetStr("log_db"))
	maps.Start(player.ENC,lib.GetServerRoot()+"/config/maps")
	// 启动控制台
	consolePort := config.Server.GetInt("console_port")
	kernel.StartConsole(map[string]kernel.ConsoleHandler{"reload": reload, "reload_config": reloadConfig}, consolePort)
	svrList := []boot.Boot{
		{service.AccountActor, global.SYS_ACCOUNT_SERVER},
	}
	boot.StartService(serviceSup, svrList)
	// 最后启动网关
	gatePort := config.Server.GetInt("game_port")
	gate.Start("gw", gw.TcpClientActor, gatePort, gate.AcceptNum(5), gate.ClientArgs{player.ENC})
	kernel.ErrorLog("start completed")
}

func stopGame() {
	gate.Stop("gw")
	time.Sleep(time.Second)
	kernel.AppStop("MapApp")
}

func reload(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	mod := req.Form.Get("mod")
	file := kernel.GetMainRoot() + "/hotfix/" + mod + "/" + mod + ".so"
	p, e := plugin.Open(file)
	if e != nil {
		kernel.ErrorLog("%s", e.Error())
		fmt.Fprintf(w, "so can not load: %s\n", file)
		return
	}
	f, err := p.Lookup("HotFix")
	if err != nil {
		kernel.ErrorLog("%s", err.Error())
		fmt.Fprintf(w, "HotFix func can not load\n")
		return
	}
	f.(func())()
	kernel.ErrorLog("reload: %s", file)
	fmt.Fprintf(w, "reload: %s\n", file)
}

func reloadConfig(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	name := req.Form.Get("name")
	var ok bool
	if name == "Server" {
		ok = config.Load(name, "./config/json/system")
	} else {
		ok = config.Load(name, "./config/json")
	}
	if ok {
		kernel.ErrorLog("reload config: %s", name)
		fmt.Fprintf(w, "reload config: %s\n", name)
	} else {
		kernel.ErrorLog("error: cannot reload config: %s", name)
		fmt.Fprintf(w, "error: cannot reload config: %s\n", name)
	}
}
