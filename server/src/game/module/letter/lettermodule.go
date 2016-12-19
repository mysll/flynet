package letter

import (
	"server"
	"server/data/datatype"
)

var Module *LetterModule

type GetContainer func(player datatype.Entity, item datatype.Entity) datatype.Entity

type LetterModule struct {
	server.BaseModule
	fc     GetContainer
	Letter *LetterSystem
}

func (m *LetterModule) Init() {
	m.Letter = NewLetterSystem()
	server.RegisterHandler("MailBox", m.Letter)
	server.RegisterCallee("Player", &PlayerLetter{})
}

func (m *LetterModule) SetContainerFindFunc(findcontainer GetContainer) {
	m.fc = findcontainer
}

func init() {
	Module = &LetterModule{}
}
