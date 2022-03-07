package maps

import (
	"game/boot"
	"game/lib"
	"github.com/liangmanlin/gootp/kernel"
	"strconv"
)

func (a *app) Name() string {
	return "MapApp"
}

func (a *app) Start(bootType kernel.AppBootType) *kernel.Pid {
	pid := kernel.SupStart("map_sup")
	boot.StartService(pid,[]boot.Boot{
		{Name: "map_agent",Svr: agentSvr},
	})
	// 加载配置
	loadConfig(a.mapConfigPath)
	// 启动场景
	for mapID := range allMaps {
		for i:=0;i<200;i++ {
			StartMap(mapID, lib.NormalMapName(mapID)+strconv.Itoa(i), ModMap["mod_common"])
		}
	}
	kernel.ErrorLog("map started")
	return pid
}

func (a *app) Stop(stopType kernel.AppStopType) {
	kernel.ErrorLog("map stopped")
}

func (a *app) SetEnv(key string, value interface{}) {

}

func (a *app) GetEnv(key string) interface{} {
	return nil
}
