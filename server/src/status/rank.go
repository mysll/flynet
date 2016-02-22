package status

import (
	"container/list"
	"fmt"
	"libs/log"
	"libs/rpc"
	"server"
	"share"
	"time"
	"util"
)

type RankInfo struct {
	Id       uint64
	Name     string
	Score    int32
	UserData []byte
	saved    bool
}

type Rank struct {
	rows      *list.List
	max_count int
	dirty     bool
	save_data time.Time
}

func (rank *Rank) FindPos(score int32) *list.Element {
	for e := rank.rows.Front(); e != nil; e = e.Next() {
		ri := e.Value.(*RankInfo)
		if ri.Score < score {
			return e
		}
	}
	return nil
}

func (rank *Rank) FindInfoById(id uint64) *RankInfo {
	for e := rank.rows.Front(); e != nil; e = e.Next() {
		ri := e.Value.(*RankInfo)
		if ri.Id == id {
			return ri
		}
	}

	return nil
}

func (rank *Rank) FindElementById(id uint64) (index int, e *list.Element) {
	index = -1
	for e = rank.rows.Front(); e != nil; e = e.Next() {
		index++
		ri := e.Value.(*RankInfo)
		if ri.Id == id {
			return index, e
		}
	}

	return -1, nil
}

func NewRank(count int) *Rank {
	return &Rank{
		rows:      list.New(),
		max_count: count,
	}
}

type RankManager struct {
	ranks map[string]*Rank
}

func NewRankManager() *RankManager {
	rm := &RankManager{}
	rm.ranks = map[string]*Rank{
		"survive_ranking": NewRank(100),
	}
	return rm
}

func (rm *RankManager) savetotal() {
	db := server.GetAppByType("database")
	if db == nil {
		log.LogError("db not found")
		return
	}
	dbwarp := server.NewDBWarp(db)
	for k, v := range rm.ranks {
		dbwarp.DeleteRow(nil, k, "", "_", share.DBParams{})
		if v.rows.Len() == 0 {
			continue
		}

		args := make([]interface{}, 0, v.rows.Len()*4)
		for e := v.rows.Front(); e != nil; e = e.Next() {
			args = append(args, e.Value.(*RankInfo).Id)
			args = append(args, e.Value.(*RankInfo).Name)
			args = append(args, e.Value.(*RankInfo).Score)
			args = append(args, e.Value.(*RankInfo).UserData)
		}

		dbwarp.InsertRows(nil, k, []string{"id", "name", "score", "userdata"}, args, "RankManager.SaveBack", share.DBParams{"rankname": k})
	}
}

func (rm *RankManager) checkexpire(intervalid server.TimerID, count int32, args interface{}) {
	db := server.GetAppByType("database")
	if db == nil {
		log.LogError("db not found")
		return
	}
	dbwarp := server.NewDBWarp(db)
	curtime := time.Now()
	for k, v := range rm.ranks {
		if !util.IsSameDay(v.save_data, curtime) { //过期删除
			var next *list.Element
			for e := v.rows.Front(); e != nil; e = next {
				next = e.Next()
				v.rows.Remove(e)
			}
			v.dirty = false
			v.save_data = curtime

			dbwarp.DeleteRow(nil, k, "", "_", share.DBParams{})
			dbwarp.UpdateRow(nil, "pub_data", map[string]interface{}{"value": curtime.Format(util.TIME_LAYOUT)}, "`key`='survive_ranking_save_data'", "_", share.DBParams{})
		}
	}
}

func (rm *RankManager) timecheck(intervalid server.TimerID, count int32, args interface{}) {
	db := server.GetAppByType("database")
	if db == nil {
		log.LogError("db not found")
		return
	}
	dbwarp := server.NewDBWarp(db)
	for k, v := range rm.ranks {
		if v.dirty {
			dbwarp.DeleteRow(nil, k, "", "_", share.DBParams{})
			args := make([]interface{}, 0, v.rows.Len()*4)
			for e := v.rows.Front(); e != nil; e = e.Next() {
				args = append(args, e.Value.(*RankInfo).Id)
				args = append(args, e.Value.(*RankInfo).Name)
				args = append(args, e.Value.(*RankInfo).Score)
				args = append(args, e.Value.(*RankInfo).UserData)
			}

			dbwarp.InsertRows(nil, k, []string{"id", "name", "score", "userdata"}, args, "RankManager.SaveBack", share.DBParams{"rankname": k})
		}
	}
}

func (rm *RankManager) SaveBack(mailbox rpc.Mailbox, args share.DBParams, eff int, err string) error {
	if err != "" {
		return fmt.Errorf(err)
	}
	rankname := args["rankname"].(string)
	var rank *Rank
	var ok bool
	if rank, ok = rm.ranks[rankname]; !ok {
		return fmt.Errorf("%s not found", rankname)
	}

	rank.dirty = false
	log.LogMessage("save ", rankname, " ok")
	return nil
}

func (rm *RankManager) load() {
	db := server.GetAppByType("database")
	if db == nil {
		log.LogError("db not found")
		return
	}

	server.NewDBWarp(db).QueryRows(nil, "survive_ranking", "", "score DESC", 0, 100, "RankManager.LoadSurvive", share.DBParams{"rankname": "survive_ranking"})
	App.AddTimer(time.Minute, -1, rm.checkexpire, nil)
	App.AddTimer(time.Minute*5, -1, rm.timecheck, nil)
}

func (rm *RankManager) LoadSurvive(mailbox rpc.Mailbox, params share.DBParams, result []share.DBRow, index, count int) error {
	if len(result) == 1 {
		if errmsg, err := result[0].GetString("error"); err == nil {
			log.LogError(errmsg)
			return nil
		}
	}
	rankname := params["rankname"].(string)

	var rank *Rank
	var ok bool
	if rank, ok = rm.ranks[rankname]; !ok {
		log.LogError(rankname, " is not found")
	}

	var next *list.Element

	for e := rank.rows.Front(); e != nil; e = next {
		next = e.Next()
		rank.rows.Remove(e)
	}

	var err error
	for _, v := range result {
		info := &RankInfo{}
		if info.Id, err = v.GetUint64("id"); err != nil {
			return nil
		}
		if info.Name, err = v.GetString("name"); err != nil {
			return nil
		}
		if info.Score, err = v.GetInt32("score"); err != nil {
			return nil
		}
		if info.UserData, err = v.GetBytes("userdata"); err != nil {
			return nil
		}
		rank.rows.PushBack(info)
	}

	log.LogMessage("load ", rankname, " ok")

	db := server.GetAppByType("database")
	if db == nil {
		log.LogError("db not found")
		return server.ErrAppNotFound
	}
	server.NewDBWarp(db).QueryRow(nil, "pub_data", "`key`='survive_ranking_save_data'", "", "RankManager.LoadPubData", share.DBParams{"key": "survive_ranking_save_data"})

	return nil
}

func (rm *RankManager) LoadPubData(mailbox rpc.Mailbox, params share.DBParams, result share.DBRow) error {

	db := server.GetAppByType("database")
	if db == nil {
		log.LogError("db not found")
		return server.ErrAppNotFound
	}

	dbwarp := server.NewDBWarp(db)

	key := params["key"].(string)
	if result.Count() == 0 {
		switch key {
		case "survive_ranking_save_data":
			return dbwarp.InsertRow(nil, "pub_data", map[string]interface{}{"key": "survive_ranking_save_data", "value": time.Now().Format(util.TIME_LAYOUT)}, "_", share.DBParams{})
		}
	}

	var rank *Rank
	switch key {
	case "survive_ranking_save_data":
		rank = rm.ranks["survive_ranking"]
		dt, err := result.GetString("value")
		if err != nil {
			log.LogError("value not found")
			break
		}
		if t, err := time.Parse(util.TIME_LAYOUT, dt); err != nil {
			rank.save_data = time.Now()
		} else {
			rank.save_data = t
		}

		if !util.IsSameDay(rank.save_data, time.Now()) { //过期删除
			var next *list.Element
			for e := rank.rows.Front(); e != nil; e = next {
				next = e.Next()
				rank.rows.Remove(e)
			}
			rank.dirty = false
			rank.save_data = time.Now()

			dbwarp.DeleteRow(nil, "survive_ranking", "", "_", share.DBParams{})
			dbwarp.UpdateRow(nil, "pub_data", map[string]interface{}{"value": time.Now().Format(util.TIME_LAYOUT)}, "`key`='survive_ranking_save_data'", "_", share.DBParams{})
			log.LogMessage("clear survive_ranking")
		}
	}

	return nil
}

func (rm *RankManager) GetRankGrade(mailbox rpc.Mailbox, rankname string, id uint64, callback string) error {
	var rank *Rank
	var ok bool
	if rank, ok = rm.ranks[rankname]; !ok {
		return fmt.Errorf("%s not found", rankname)
	}

	index, _ := rank.FindElementById(id)

	app := server.GetApp(mailbox.Address)
	if app == nil {
		return server.ErrAppNotFound
	}

	return app.Call(&mailbox, callback, rankname, id, index)
}

func (rm *RankManager) UpdateUserInfo(mailbox rpc.Mailbox, rankname string, id uint64, name string, score int32, userdata []byte) error {
	db := server.GetAppByType("database")
	if db == nil {
		return fmt.Errorf("db not found")
	}

	var rank *Rank
	var ok bool
	if rank, ok = rm.ranks[rankname]; !ok {
		return fmt.Errorf("%s not found", rankname)
	}

	count := rank.rows.Len()
	if count == 0 { //榜单为空
		rank.rows.PushBack(&RankInfo{id, name, score, userdata, false})
		rank.dirty = true
		return nil
	}

	pos := rank.FindPos(score)

	if count < rank.max_count { //还有空的位置
		if pos == nil {
			rank.rows.PushBack(&RankInfo{id, name, score, userdata, false})
		} else {
			rank.rows.InsertBefore(&RankInfo{id, name, score, userdata, false}, pos)
		}
		rank.dirty = true
	}

	if pos == nil { //没有进入榜单
		return nil
	}

	//插入榜单
	rank.rows.InsertBefore(&RankInfo{id, name, score, userdata, false}, pos)

	//删除最后一个
	ele := rank.rows.Back()

	rank.rows.Remove(ele)
	rank.dirty = true
	return nil
}

func (rm *RankManager) DelBack(mailbox rpc.Mailbox, params share.DBParams, eff int, err string) error {
	if err != "" {
		return fmt.Errorf(err)
	}
	log.LogMessage("delete from ", params["rankname"].(string), " id = ", params["id"].(uint64))
	return nil
}

func (rm *RankManager) InsertBack(mailbox rpc.Mailbox, params share.DBParams, eff int, err string) error {
	if err != "" {
		return fmt.Errorf(err)
	}
	var rank *Rank
	switch params["rankname"].(string) {
	case "survive_ranking":
		rank = rm.ranks["survive_ranking"]
	}

	if rank == nil {
		return fmt.Errorf("rank not found")
	}

	info := rank.FindInfoById(params["id"].(uint64))
	if info == nil {
		return fmt.Errorf("id not found")
	}

	info.saved = true
	return nil
}
