package global

const (
	DB_OP_ADD int32 = iota + 1
	DB_OP_UPDATE
	DB_OP_DELETE
)

const GOODS_TYPE_DIV int32 = 10000000

const (
	GOODS_TYPE_ITEM int32 = iota + 1 // 普通道具
)

const (
	BACKUP_BAG_ID int32 = iota
	BACKUP_BAG_INSERT
	BACKUP_BAG_UPDATE_NUM
	BACKUP_BAG_DELETE
	BACKUP_ATTR
)



