package basemgr

import (
	"server"
	"server/libs/log"
	"server/libs/rpc"
	"sort"
	"sync"
)

type Session struct {
	l      sync.Mutex
	id     uint64
	serial int
}

func (t *Session) RegisterCallback(s rpc.Servicer) {
	s.RegisterCallback("GetBaseAndId", t.GetBaseAndId)
}

func (s *Session) GetBaseAndId(mailbox rpc.Mailbox, msg *rpc.Message) (errcode int32, reply *rpc.Message) {
	r := server.NewMessageReader(msg)
	user, err := r.ReadString()
	if server.Check(err) {
		return 0, nil
	}
	s.l.Lock()
	defer s.l.Unlock()

	s.serial++
	if s.serial < 0 {
		s.serial = 0
	}
	bases := server.GetAppIdsByType("base")
	sort.Sort(sort.StringSlice(bases))
	if len(bases) > 0 {
		idx := s.serial % len(bases)
		baseid := bases[idx]
		s.id++
		if base := server.GetAppByName(baseid); base != nil {
			server.Check(base.Call(&mailbox, "Login.AddClient", user))
			return 0, nil
		}

		log.LogError(server.ErrNotFoundApp)
		return 0, nil
	}

	log.LogError(server.ErrNotFoundApp)
	return 0, nil
}
