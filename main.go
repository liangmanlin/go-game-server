package main

import (
	"fmt"
	"game/config"
	"game/global"
	"game/gw"
	"game/player"
	"game/proto"
	"game/proto/router"
	"github.com/liangmanlin/gootp/db"
	"github.com/liangmanlin/gootp/gate"
	"github.com/liangmanlin/gootp/gate/pb"
	"github.com/liangmanlin/gootp/kernel"
	"github.com/liangmanlin/gootp/node"
	"net/http"
	"plugin"
)

func main() {
	config.LoadAll("./config/json")
	kernel.Env.WriteLogStd = true
	kernel.Env.LogPath = config.Server.GetStr("logs_dir")
	kernel.Env.SetTimerMinTick(100)
	// 设置Env需要在启动之前
	kernel.KernelStart(initGame, stopGame)
}

func initGame() {
	node.Env.PingTick = 30*1000 // 可以指定心跳间隔,默认60*1000ms
	nodeProto := []interface{}{&global.TestCall{},&global.StrArgs{},&global.RpcStrResult{}}
	Cookie := config.Server.GetListStr("node",1)
	node.Start(config.Server.GetListStr("node",0), Cookie, nodeProto)
	kernel.SupStartChild("kernel", &kernel.SupChild{ChildType: kernel.SupChildTypeSup, Name: "gameSup"})
	player.ENC = pb.Parse(proto.TOC, 0)
	player.DEC = pb.Parse(proto.TOS, 1)
	// 删除没有的数据，以免浪费gc的时间
	proto.TOC = nil
	proto.TOS = nil
	player.Router = router.MakeRouter()
	dbConfig := config.Server.Get("db").([]interface{})
	port :=int(dbConfig[1].(float64))
	DBConfig := db.Config{Host: dbConfig[0].(string), Port: port, User: dbConfig[2].(string), PWD: dbConfig[3].(string)}
	db.Start(DBConfig,global.DBTables, config.Server.GetStr("game_db"), config.Server.GetStr("log_db"))
	gatePort := config.Server.GetInt("game_port")
	gate.Start("gw", gw.TcpClientActor, gatePort, gate.AcceptNum(5), gate.ClientArgs{player.ENC})
	// 启动控制台
	consolePort := config.Server.GetInt("console_port")
	kernel.StartConsole(map[string]kernel.ConsoleHandler{"reload": reload,"reload_config":reloadConfig}, consolePort)
	//actor := kernel.DefaultActor()
	//actor.HandleCast = func(context *kernel.Context, msg interface{}) {
	//	kernel.ErrorLog("rec cast:%#v", msg)
	//}
	//actor.HandleCall = func(context *kernel.Context, request interface{}) interface{} {
	//	kernel.ErrorLog("rec call: %#v", request)
	//	return request
	//}
	//kernel.StartName("ttt", actor)
	//f := func(i interface{})interface {}{kernel.ErrorLog("%#v",i);return i}
	//node.RpcRegister("Echo", &f)
	//kernel.Start(player.Actor,nil,nil,int64(1))
	kernel.ErrorLog("start completed")
}

func stopGame() {
	gate.Stop()
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

func reloadConfig(w http.ResponseWriter, req *http.Request)  {
	req.ParseForm()
	name := req.Form.Get("name")
	var ok bool
	if name == "Server" {
		ok = config.Load(name,"./config/json/system")
	}else {
		ok = config.Load(name,"./config/json")
	}
	if ok {
		kernel.ErrorLog("reload config: %s", name)
		fmt.Fprintf(w, "reload config: %s\n", name)
	}else {
		kernel.ErrorLog("error: cannot reload config: %s", name)
		fmt.Fprintf(w, "error: cannot reload config: %s\n", name)
	}
}