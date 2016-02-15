package basemgr

import (
	"libs/rpc"
	"server"
	"sort"
	"sync"
)

type Session struct {
	l      sync.Mutex
	id     uint64
	serial int
}

func (s *Session) GetBaseAndId(mailbox rpc.Mailbox, user string) error {
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
		if base := server.GetApp(baseid); base != nil {
			return base.Call(&mailbox, "Login.AddClient", user)
		}

		return server.ErrNotFoundApp
	}

	return server.ErrNotFoundApp
}
