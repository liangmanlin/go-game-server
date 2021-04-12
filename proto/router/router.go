package router

import "game/global"
import "game/player"

func MakeRouter()map[int]*global.HandleFunc{
	var rs = map[int]*global.HandleFunc{
		101:&player.LoginLogin,
		201:&player.LoginLogin,
	}
	return rs
}