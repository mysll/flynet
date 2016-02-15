package mongodb

import (
	"data/entity"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"share"
	"strings"
)

func SaveToDb(db *mgo.Database, id uint64, savedata *share.DbSave) error {
	c := db.C(strings.ToLower(savedata.Data.Typ))
	savedata.Data.DBId = id
	_, err := c.Upsert(bson.M{"_id": id}, *savedata.Data)
	return err
}

func SaveItem(db *mgo.Database, id uint64, data *share.SaveEntity, pname string, pid uint64, index int) error {
	var err error
	if data == nil {
		return nil
	}

	if data.DBId == 0 {
		data.DBId = id
	}

	c := db.C(strings.ToLower(data.Typ))
	err = c.FindId(data.DBId).One(nil)
	if err != nil { //没有找到
		if err = c.Insert(bson.M{"_id": data.DBId}); err != nil {
			return err
		}
	}

	err = c.Update(bson.M{"_id": data.DBId}, data.Obj)

	if err != nil {
		return err
	}

	//基类数据
	if data.Base != nil {
		err = SaveItem(db, data.DBId, data.Base, "", 0, -1)
		if err != nil {
			return err
		}
	}

	if pname != "" {
		c1 := db.C(strings.ToLower(pname) + "_child")
		cinfo := Childs{
			Parent_Id: pid,
			Child_Id:  data.DBId,
			Type:      data.Typ,
			Index:     index,
		}

		if err := c1.Insert(cinfo); err != nil {
			return err
		}
	}

	c1 := db.C(strings.ToLower(data.Typ) + "_child")
	c1.Remove(bson.M{"parent_id": data.DBId})

	for _, e := range data.Childs {
		err = SaveItem(db, 0, e, data.Typ, data.DBId, data.Index)
		if err != nil {
			return err
		}
	}

	return nil
}

func LoadUser(db *mgo.Database, uid uint64, ent string) (loaddata share.DbSave, err error) {
	loaddata.Data, err = LoadEntity(db, uid, ent, 0)
	return
}

func parseEntity(data *share.SaveEntity) error {

	if data == nil {
		return nil
	}

	bdata := data.Obj.(bson.M)
	e := entity.CreateSaveLoader(data.Typ)
	for k, v := range bdata {
		if strings.HasSuffix(k, "_save_property") {
			prop := v.(bson.M)
			for k1, v1 := range prop {
				bdata[k1] = v1
			}
			delete(bdata, k)
			break
		}
	}

	err := e.Unmarshal(bdata)
	if err != nil {
		return err
	}

	data.Obj = e
	if data.Base != nil {
		err = parseEntity(data.Base)
		if err != nil {
			return err
		}
	}

	for _, c := range data.Childs {
		err = parseEntity(c)
		if err != nil {
			return err
		}
	}

	return nil

}

func LoadEntity(db *mgo.Database, uid uint64, ent string, index int) (*share.SaveEntity, error) {
	si := &share.SaveEntity{}

	c := db.C(strings.ToLower(ent))
	err := c.Find(bson.M{"_id": uid}).One(si)
	if err == nil {
		err = parseEntity(si)
	}
	return si, err
}
