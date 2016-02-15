package server

const (
	NEWUSERCONN  = "new_user_conn"
	LOSTUSERCONN = "lost_user_conn"
	MASERTINMSG  = "master_inmsg"
	NEWAPPREADY  = "new_app_ready"
	APPLOST      = "app_lost"
)

type MasterMsg struct {
	Id   uint16
	Body []byte
}
