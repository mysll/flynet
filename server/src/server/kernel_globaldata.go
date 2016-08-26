package server

import "server/libs/rpc"

type GlobalDataHelper struct {
}

func (gd *GlobalDataHelper) RegisterCallback(s rpc.Servicer) {
	s.RegisterCallback("SyncGlobalData", gd.SyncGlobalData)
}

func (gd *GlobalDataHelper) SyncGlobalData(sender rpc.Mailbox, msg *rpc.Message) *rpc.Message {
	return nil
}
