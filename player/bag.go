package player

import (
	"game/global"
	"game/lib"
	"github.com/liangmanlin/gootp/db"
	"github.com/liangmanlin/gootp/gutil"
	"github.com/liangmanlin/gootp/kernel"
)

const bagName = "bag"

var BagLoad = func(ctx *kernel.Context, player *global.Player) {
	rl := db.SyncSelect(ctx, "bag", player.RoleID, player.RoleID)
	m := make(map[int32]*global.PGoods)
	var maxID int32
	tm := make(map[int32][]int32)
	for _, v := range rl {
		goods := toPGoods(v.(*global.Bag))
		m[goods.ID] = goods
		maxID = gutil.MaxInt32(maxID, goods.ID)
		tmp := tm[goods.TypeID]
		tm[goods.TypeID] = append(tmp, goods.ID)
	}
	player.BagMaxID = maxID // todo 这个值可以考虑持久化
	player.Bag = &global.BagData{MaxSize: 100, Goods: m, TypeIDMap: tm}
}

var BagPersistent = func(ctx *kernel.Context, player *global.Player) {
	if len(player.Bag.Dirty) > 0 {
		for id, v := range player.Bag.Dirty {
			switch v.Type {
			case global.DB_OP_ADD:
				db.SyncInsert("bag", player.RoleID, toBag(player.Bag.Goods[id]))
			case global.DB_OP_UPDATE:
				db.SyncUpdate("bag", player.RoleID, toBag(player.Bag.Goods[id]))
			case global.DB_OP_DELETE:
				db.SyncDelete("bag", player.RoleID, toBag(v.Goods))
			}
		}
		player.Bag.Dirty = make(map[int32]global.GoodsDirty)
	}
}

var BagGetGoodsByID = func(player *global.Player, id int32) *global.PGoods {
	if g, ok := player.Bag.Goods[id]; ok {
		return g
	}
	return nil
}

var BagGetGoodsByTypeID = func(player *global.Player, typeID int32) (goodsList []*global.PGoods) {
	tmp := player.Bag.TypeIDMap[typeID]
	for _, id := range tmp {
		goodsList = append(goodsList, player.Bag.Goods[id])
	}
	return
}

// 建议使用该函数获取道具数量判断
var BagGetGoodsNum = func(player *global.Player, typeID int32) (notBindNum, bindNum int32) {
	tmp := player.Bag.TypeIDMap[typeID]
	for _, id := range tmp {
		g := player.Bag.Goods[id]
		if g.Bind {
			bindNum += g.Num
		} else {
			notBindNum += g.Num
		}
	}
	return
}

var BagCreateGoods = func(player *global.Player, typeID, num int32, bind bool) {
	pileNum := lib.GetPileNum(typeID)
	// 先判断是否可以堆叠
	if pileNum > 1 {
		BagInsertGoodsTypeID(player, typeID, num, bind, pileNum)
	} else {
		// 不可堆叠,直接创建插入
		goodsList := lib.CreateGoods(typeID, num, bind)
		BagInsertGoodsNotPile(player, goodsList)
	}
}

var BagInsertGoodsNotPile = func(player *global.Player, goodsList []*global.PGoods) {
	checkTransaction(player)
	// 预先判断减少回滚消耗
	if int(player.Bag.MaxSize)-len(player.Bag.Goods) < len(goodsList) {
		Abort(&global.PMsg{})
	}
	for _, goods := range goodsList {
		goods.RoleID = player.RoleID
		goods.ID = bagMakeID(player)
		bagAddGoods(player, goods)
	}
}

var BagInsertGoodsTypeID = func(player *global.Player, typeID, num int32, bind bool, pileNum int32) {
	checkTransaction(player)
	bag := player.Bag
	tm := bag.TypeIDMap[typeID]
	for i := 0; i < len(tm) && num > 0; i++ {
		goods := bag.Goods[tm[i]]
		if bind == goods.Bind && goods.Num < pileNum {
			add := gutil.MinInt32(pileNum-goods.Num, num)
			num -= add
			Backup(player, global.BackupKey{BackupID: global.BACKUP_BAG_UPDATE_NUM, ID: goods.ID}, rollbackBagGoodsNum, []int32{goods.ID, goods.Num})
			goods.Num += add
			AddDBQueue(player, BagDBOP, global.DB_OP_UPDATE, goods.ID)
		}
	}
	if num > 0 {
		goodsList := lib.CreateGoods(typeID, num, bind)
		BagInsertGoodsNotPile(player, goodsList)
	}
}

// 上层逻辑不建议直接修改数据，包括数量,除非你清楚自己在做什么
// 上层逻辑如果自己修改了PGoods的数据，需要自行调用，否则不会更新到数据库
// 原则不提倡自行修改，因为事务也会失效,所以只能在事务不会回滚的地方使用
var BagUpdateGoods = func(player *global.Player, id int32) {
	AddDBQueue(player, BagDBOP, global.DB_OP_UPDATE, id)
}

// 优先扣除绑定的道具
var BagDeleteByTypeID = func(player *global.Player, typeID, num int32) {
	checkTransaction(player)
	bag := player.Bag
	tm := bag.TypeIDMap[typeID]
	var bind, notBind []*global.PGoods
	var total int32
	for _, id := range tm {
		g := player.Bag.Goods[id]
		total += g.Num
		if g.Bind {
			bind = append(bind, g)
		} else {
			notBind = append(notBind, g)
		}
	}
	if total < num {
		Abort(&global.PMsg{})
	}
	// 走到这里肯定是足够的
	num = deleteList(player, bind, num)
	if num > 0 {
		deleteList(player, notBind, num)
	}
}

var BagDeleteByID = func(player *global.Player, id, num int32) {
	checkTransaction(player)
	bag := player.Bag
	goods := bag.Goods[id]
	if goods.Num < num {
		Abort(&global.PMsg{})
	}
	if num == goods.Num {
		bagDeleteGoods(player,goods)
	}else{
		Backup(player, global.BackupKey{BackupID: global.BACKUP_BAG_UPDATE_NUM, ID: goods.ID}, rollbackBagGoodsNum, []int32{goods.ID, goods.Num})
		AddDBQueue(player, BagDBOP, global.DB_OP_UPDATE, goods.ID)
		goods.Num -= num
	}
}

var BagDBOP = func(player *global.Player, op int32, arg interface{}) {
	bag := player.Bag
	switch op {
	case global.DB_OP_ADD:
		id := arg.(int32)
		bag.Dirty[id] = global.GoodsDirty{Type: op}
	case global.DB_OP_UPDATE:
		id := arg.(int32)
		if _, ok := bag.Dirty[id]; !ok {
			bag.Dirty[id] = global.GoodsDirty{Type: op}
		}
	case global.DB_OP_DELETE:
		goods := arg.(*global.PGoods)
		if f, ok := bag.Dirty[goods.ID]; ok {
			if f.Type == global.DB_OP_ADD {
				delete(bag.Dirty, goods.ID)
			} else {
				bag.Dirty[goods.ID] = global.GoodsDirty{Type: op, Goods: goods}
			}
		} else {
			bag.Dirty[goods.ID] = global.GoodsDirty{Type: op, Goods: goods}
		}
	}
	player.DirtyMod[bagName] = true
}

func deleteList(player *global.Player, arr []*global.PGoods, num int32) int32 {
	for i := 0; i < len(arr) && num > 0; i++ {
		g := arr[i]
		if num >= g.Num {
			bagDeleteGoods(player,g)
			num -= g.Num
		}else {
			Backup(player, global.BackupKey{BackupID: global.BACKUP_BAG_UPDATE_NUM, ID: g.ID}, rollbackBagGoodsNum, []int32{g.ID, g.Num})
			AddDBQueue(player, BagDBOP, global.DB_OP_UPDATE, g.ID)
			g.Num -= num
			num = 0
		}
	}
	return num
}

func bagDeleteGoods(player *global.Player,goods *global.PGoods)  {
	Backup(player, global.BackupKey{BackupID: global.BACKUP_BAG_DELETE, ID: goods.ID}, rollbackBagDelete, goods)
	AddDBQueue(player, BagDBOP, global.DB_OP_DELETE, kernel.DeepCopy(goods))
	bag := player.Bag
	delete(bag.Goods,goods.ID)
	if tl,ok := bag.TypeIDMap[goods.TypeID];ok{
		bag.TypeIDMap[goods.TypeID] = gutil.SliceDelInt32(tl, goods.ID)
	}

}

// 直接插入
func bagAddGoods(player *global.Player, goods *global.PGoods) {
	Backup(player, global.BackupKey{BackupID: global.BACKUP_BAG_INSERT, ID: goods.ID}, rollbackBagInsert, goods.ID)
	AddDBQueue(player, BagDBOP, global.DB_OP_ADD, goods.ID)
	bag := player.Bag
	bag.Goods[goods.ID] = goods
	arr := bag.TypeIDMap[goods.TypeID]
	bag.TypeIDMap[goods.TypeID] = append(arr, goods.ID)
}

func rollbackBagInsert(player *global.Player, id int32) {
	bag := player.Bag
	if goods, ok := bag.Goods[id]; ok {
		delete(bag.Goods, id)
		if arr, ok := bag.TypeIDMap[goods.TypeID]; ok {
			bag.TypeIDMap[goods.TypeID] = gutil.SliceDelInt32(arr, id)
		}
	}
}

func rollbackBagDelete(player *global.Player,goods *global.PGoods)  {
	bag := player.Bag
	bag.Goods[goods.ID] = goods
	arr := bag.TypeIDMap[goods.TypeID]
	bag.TypeIDMap[goods.TypeID] = append(arr, goods.ID)
}

func rollbackBagGoodsNum(player *global.Player, arg []int32) {
	id, num := arg[0], arg[1]
	bag := player.Bag
	if goods, ok := bag.Goods[id]; ok {
		goods.Num = num
	}
}

func bagMakeID(player *global.Player) int32 {
	Backup(player, global.BackupKey{BackupID: global.BACKUP_BAG_ID}, "BagMaxID", player.BagMaxID)
	player.BagMaxID++
	return player.BagMaxID
}

func toPGoods(bagGoods *global.Bag) *global.PGoods {
	goods := &global.PGoods{
		RoleID:     bagGoods.RoleID,
		ID:         bagGoods.ID,
		Type:       bagGoods.Type,
		TypeID:     bagGoods.TypeID,
		Num:        bagGoods.Num,
		Bind:       bagGoods.Bind,
		StartTime:  bagGoods.StartTime,
		EndTime:    bagGoods.EndTime,
		CreateTime: bagGoods.CreateTime,
	}
	return goods
}

func toBag(bagGoods *global.PGoods) *global.Bag {
	goods := &global.Bag{
		RoleID:     bagGoods.RoleID,
		ID:         bagGoods.ID,
		Type:       bagGoods.Type,
		TypeID:     bagGoods.TypeID,
		Num:        bagGoods.Num,
		Bind:       bagGoods.Bind,
		StartTime:  bagGoods.StartTime,
		EndTime:    bagGoods.EndTime,
		CreateTime: bagGoods.CreateTime,
	}
	return goods
}
