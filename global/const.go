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

const ROLEID_AGENT int64 = 1000000000000
const ROLEID_SERVER int64 = 1000000

const (
	SYS_ACCOUNT_SERVER = "sys_account_server"
)

const (
	ACTOR_ROLE int8 = iota + 1
	ACTOR_MONSTER
	C_ACTOR_SIZE // 这个一动要在最后
)

const CHECK_NODE_PROTO = true
