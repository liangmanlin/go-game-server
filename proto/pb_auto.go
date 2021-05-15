package proto

import "game/global"

// TODO 自动生成，请勿手工修改

var TOS = map[int]interface{}{
	101:&global.LoginTosConnect{},
	102:&global.LoginTosLogin{},
	103:&global.LoginTosCreateRole{},
	104:&global.LoginTosSelect{},
	201:&global.GameTosUP{},
	306:&global.MapTosMove{},
	307:&global.MapTosMoveStop{},
}

var TOC = map[int]interface{}{
	101:&global.LoginTocConnect{},
	102:&global.LoginTocLogin{},
	103:&global.LoginTocCreateRole{},
	104:&global.LoginTocSelect{},
	201:&global.GameTocUP{},
	301:&global.MapTocActorLeaveArea{},
	302:&global.MapTocEnterArea{},
	303:&global.MapTocRoleEnterArea{},
	304:&global.MapTocMonsterEnterArea{},
	305:&global.MapTocStop{},
	306:&global.MapTocMove{},
	307:&global.MapTocMoveStop{},
	308:&global.MapTocUpdateMonsterInfo{},
	309:&global.MapTocUpdateRoleInfo{},
	310:&global.MapTocActorDead{},
	311:&global.MapTocDelBuff{},
	401:&global.FightTocUseSkill{},
	402:&global.FightTocSkillEffect{},
	501:&global.RoleTocUPProps{},
}