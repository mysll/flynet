package share

type CreateItemArgs struct {
	Itemid string
	Amount int16
}

const (
	COMMANG_NONE      = 1000 + iota
	PLAYER_FIRST_LAND //玩家第一次进游戏
	COMMAND_CUSTOM    //自定义消息起始消息号
)
