package mongodb

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"libs/log"
	//. "libs/parser"
	//"os"
	"server"
	//"strings"
	"util"
)

var (
	ROLEINFO = "role_info"
	COUNTERS = "counters"
	db       *MongoDB
)

type MongoDB struct {
	session *mgo.Session
	DB      *mgo.Database
	wg      util.WaitGroupWrapper
	Account *Account
	pools   int
	limit   int
}

func (self *MongoDB) InitDB(db string, source string, threads int, entity string, role string, limit int) error {
	session, err := mgo.Dial(source)
	if err != nil {
		return err
	}
	self.session = session
	self.session.SetMode(mgo.Monotonic, true)
	self.DB = self.session.DB(db)
	self.pools = threads
	self.Account = NewAccount(threads)
	self.Account.Do()
	self.limit = limit
	server.RegisterRemote("Account", self.Account)
	self.CheckDb(entity)
	log.LogMessage("connect to mongodb:", source)
	return nil
}

func (self *MongoDB) KeepAlive() {
}

func (self *MongoDB) CheckDb(entity string) {
	if cs, _ := self.DB.CollectionNames(); len(cs) > 0 {
		return
	}

	c := self.DB.C(COUNTERS)
	seq := Counter{}
	seq.Id_ = "userid"
	seq.Seq = 10000
	c.Insert(seq)
}

func (self *MongoDB) Close() {
	self.Account.quit = true
	self.wg.Wait()
	self.session.Close()
}

func (self *MongoDB) getNextSequence(name string) uint64 {
	c := self.DB.C(COUNTERS)
	change := mgo.Change{
		Update:    bson.M{"$inc": bson.M{"seq": 1}},
		ReturnNew: true,
	}

	seq := Counter{}
	c.Find(bson.M{"_id": name}).Apply(change, &seq)
	return seq.Seq

}
func NewMongoDB() *MongoDB {
	if db != nil {
		return db
	}
	db = &MongoDB{}
	return db
}
