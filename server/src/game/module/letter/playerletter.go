package letter

import (
	"logicdata/entity"
	"server"
	"server/data/datatype"
	"server/libs/rpc"
	"server/share"
	"time"
)

type PlayerLetter struct {
	server.Callee
}

func (pl *PlayerLetter) OnReady(self datatype.Entity, first bool) int {
	if first {
		Module.GetCore().AddHeartbeat(self, "CheckLetter", time.Minute, -1, nil)

		//清理过期的邮件
		DeleteExpiredLetter(self.(*entity.Player))
	}

	return 1
}

func (pl *PlayerLetter) OnTimer(self datatype.Entity, beat string, count int32, args interface{}) int {
	switch beat {
	case "CheckLetter":
		//清理过期的邮件
		DeleteExpiredLetter(self.(*entity.Player))
		db := server.GetAppByType("database")
		if db != nil {
			server.NewDBWarp(db).LookLetter(
				nil,
				self.GetDbId(),
				"DbBridge.LookLetterBack",
				share.DBParams{"mailbox": rpc.NewMailBoxFromUid(self.UID())},
			)
		}
		return 0
	}
	return 1
}
