package share

const (
	S2C_ERROR = 10000 + iota
	S2C_LOGININFO
	S2C_LOGINSUCCEED
	S2C_ENTERBASEERR
	S2C_ROLEINFO
	S2C_RPC
)

const (
	ERROR_SUCCESS = iota
	ERROR_MSG_ILLEGAL
	ERROR_NOBASE
	ERROR_LOGIN_FAILED         //登录失败
	ERROR_LOGIN_TRY_MAX        //登录失败超过最大次数
	ERROR_SYSTEMERROR          //系统错误
	ERROR_BASE_KEY_EXPIRED     //密钥过期
	ERROR_SELECT_ROLE_ERROR    //选择角色出错
	ERROR_CREATE_ROLE_ERROR    //创建角色出错
	ERROR_ROLE_USED            //角色正在使用中
	ERROR_ROLE_ENTERAREA_ERROR //进入场景失败
	ERROR_NAME_CONFLIT         //名字冲突
	ERROR_ROLE_LIMIT           //超出人物个数限制
	ERROR_ROLE_REPLACE         //已经被顶替
	ERROR_CONTAINER_FULL       //背包满了
	ERROR_NOLOGIN              //没有登录服务器
)

//错误消息起始编号
const (
	ERROR_BATTLE  = 10000
	ERROR_PLAYER  = 11000
	ERROR_MALL    = 12000
	ERROR_COMPOSE = 13000
	ERROR_LETTER  = 14000
)

type S2CMsg struct {
	Sender string
	To     int64
	Method string
	Data   []byte
}

type S2CBrocast struct {
	Sender string
	To     []int64
	Method string
	Data   []byte
}

type S2SBrocast struct {
	To     []int64
	Method string
	Args   interface{}
}
