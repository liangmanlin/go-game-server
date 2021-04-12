package lib

import (
	"game/config"
	"game/global"
	"github.com/liangmanlin/gootp/kernel"
)

var CreateGoods = func(typeID, num int32, bind bool) []*global.PGoods {
	Type := typeID / global.GOODS_TYPE_DIV
	switch Type {
	case global.GOODS_TYPE_ITEM:
		return CreateItem(typeID, num, bind)
	}
	kernel.ErrorLog("unknow typeid:%d", typeID)
	return nil
}

var CreateItem = func(typeID, num int32, bind bool) []*global.PGoods {
	def := config.Goods.Get(typeID)
	less := num % def.Pile_num
	var goodsList []*global.PGoods
	now := int32(kernel.Now())
	if less > 0 {
		goodsList = append(goodsList, &global.PGoods{
			Type:       global.GOODS_TYPE_ITEM,
			TypeID:     typeID,
			Num:        less,
			Bind:       bind,
			CreateTime: now,
		})
	}
	div := num / def.Pile_num
	for i := int32(0); i < div; i++ {
		goodsList = append(goodsList, &global.PGoods{
			Type:       global.GOODS_TYPE_ITEM,
			TypeID:     typeID,
			Num:        def.Pile_num,
			Bind:       bind,
			CreateTime: now,
		})
	}
	return goodsList
}

var GetPileNum = func(typeID int32) int32 {
	switch typeID / global.GOODS_TYPE_DIV {
	default:
		def := config.Goods.Get(typeID)
		return def.Pile_num
	}
}
