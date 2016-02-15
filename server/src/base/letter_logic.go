package base

import (
	"data/entity"
	"encoding/json"
	"libs/log"
	"libs/rpc"
	"pb/c2s"
	"server"
	"share"
	"time"
)

type Appendix struct {
	Configid   string
	UID        uint64
	Amount     int16
	RemainTime int32
}

const (
	ERR_MAILBOX_FULL = share.ERROR_LETTER + iota
	ERR_APPENDIX_NOT_EXIST
)

type LetterSystem struct {
}

func NewLetterSystem() *LetterSystem {
	return &LetterSystem{}
}

//清理过期的邮件
func DeleteExpiredLetter(player *entity.Player) {
	now := time.Now()
	rows := player.MailBox_r.GetRows()
	for i := rows - 1; i >= 0; i-- {
		st, _ := player.MailBox_r.GetSendTime(i)
		if now.Sub(time.Unix(st, 0)).Hours() >= 168.0 { //超过七天
			player.MailBox_r.Del(i)
		}
	}
}

//删除所有信件
func (l *LetterSystem) DeleteAllLetter(mailbox rpc.Mailbox, void c2s.Void) error {
	p := App.Players.GetPlayer(mailbox.Id)
	if p == nil {
		log.LogError("player not found, id:", mailbox.Id)
		//角色没有找到
		return nil
	}
	player := p.Entity.(*entity.Player)
	player.MailBox_r.Clear()
	return nil
}

//删除信件
func (l *LetterSystem) DeleteLetter(mailbox rpc.Mailbox, args c2s.Reqoperatemail) error {
	p := App.Players.GetPlayer(mailbox.Id)
	if p == nil {
		log.LogError("player not found, id:", mailbox.Id)
		//角色没有找到
		return nil
	}
	player := p.Entity.(*entity.Player)

	if len(args.Mails) == 0 {
		return nil
	}

	for _, sno := range args.Mails {
		row := player.MailBox_r.FindSerial_no(uint64(sno))
		if row == -1 {
			return nil
		}
		player.MailBox_r.Del(row)
	}
	return nil
}

//接收附件
func (l *LetterSystem) RecvAppendix(mailbox rpc.Mailbox, args c2s.Reqoperatemail) error {
	p := App.Players.GetPlayer(mailbox.Id)
	if p == nil {
		log.LogError("player not found, id:", mailbox.Id)
		//角色没有找到
		return nil
	}
	player := p.Entity.(*entity.Player)

	if player.MailBox_r.GetRows() == 0 {
		return nil
	}

	var mails []uint64
	if len(args.Mails) == 0 {
		mails = make([]uint64, 0, player.MailBox_r.GetRows())
		for i := 0; i < player.MailBox_r.GetRows(); i++ {
			sno, _ := player.MailBox_r.GetSerial_no(i)
			mails = append(mails, sno)
		}
	} else {
		mails = []uint64{uint64(args.Mails[0])}
	}

	for _, serial_no := range mails {
		row := player.MailBox_r.FindSerial_no(serial_no)
		if row == -1 {
			continue
		}
		info, err := player.MailBox_r.GetAppendix(row)
		if err != nil || info == "" {
			continue
		}

		var appendixs []Appendix
		err = json.Unmarshal([]byte(info), &appendixs)
		if err != nil {
			log.LogError(err)
			continue
		}

		flag := false
		index := -1
		var res int32
		for k, appendix := range appendixs {
			item, err := App.Kernel.CreateFromConfig(appendix.Configid)
			if err != nil { //物品不存在
				log.LogError("appendix not found ", appendix.Configid)
				continue
			}

			item.SetDbId(appendix.UID)
			switch inst := item.(type) {
			case *entity.Item:
				inst.SetAmount(appendix.Amount)
				inst.SetTime(appendix.RemainTime)
			}

			container := GetContainer(player, item)
			if container == nil {
				App.Kernel.Destroy(item.GetObjId())
				flag = true
				res = share.ERROR_SYSTEMERROR
				break
			}

			_, err = App.Kernel.AddChild(container.GetObjId(), item.GetObjId(), -1)
			if err != nil {
				App.Kernel.Destroy(item.GetObjId())
				flag = true
				res = share.ERROR_CONTAINER_FULL
				break
			}
			index = k
		}

		if flag {
			appendixs = appendixs[index+1:]
			data, _ := json.Marshal(appendixs)
			player.MailBox_r.SetAppendix(row, string(data))
			server.Error(nil, &mailbox, "Mail.Error", res)
			continue
		}

		player.MailBox_r.SetAppendix(row, "")
	}

	return nil
}

//读信件
func (l *LetterSystem) ReadLetter(mailbox rpc.Mailbox, args c2s.Reqoperatemail) error {
	p := App.Players.GetPlayer(mailbox.Id)
	if p == nil {
		log.LogError("player not found, id:", mailbox.Id)
		//角色没有找到
		return nil
	}
	player := p.Entity.(*entity.Player)

	if len(args.Mails) == 0 {
		return nil
	}

	row := player.MailBox_r.FindSerial_no(uint64(args.Mails[0]))
	if row == -1 {
		return nil
	}
	player.MailBox_r.SetIsRead(row, 1)
	return nil
}
