package helper

import (
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"
	. "server/data/datatype"
	"server/libs/log"
	"strconv"
	"strings"
)

var (
	cachefiles = make(map[string]*csvparser)
)

type PropSeter interface {
	PropertyType(p string) (int, string, error)
	Set(p string, v interface{}) error
}

func GetConfigIds(filename string) []string {
	if f, ok := cachefiles[filename]; ok {
		return f.GetIds()
	}
	return nil
}

func LoadAllConfig(path string) error {
	path = path + "/config/"
	log.LogMessage("load config from path:", path)
	fs, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}

	for _, f := range fs {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".csv") {
			p := NewParser()
			err = p.Parse(path + f.Name())
			if err != nil {
				log.LogError(err)
				continue
			}
			log.LogMessage("load config file:", f.Name(), ", ok")
			cachefiles[f.Name()] = p
		}
	}

	return nil
}

func ReadValueByKey(filename, id, key string) (string, error) {
	if f, ok := cachefiles[filename]; ok {
		index := f.GetKeyIndex(key)
		if index == -1 {
			return "", errors.New("key not found")
		}

		if infos := f.Find(id); infos != nil {
			return infos[index], nil
		}

		return "", errors.New("id not found")
	}

	return "", errors.New("file not found")
}

func ReadRowByKey(filename, id string) ([]string, error) {
	if f, ok := cachefiles[filename]; ok {
		if infos := f.Find(id); infos != nil {
			return infos, nil
		}
		return nil, errors.New("id not found")
	}

	return nil, errors.New("file not found")
}

func GetPropOpt(id string, prop string) []PropOp {
	var f *csvparser
	var infos []PropOp
	for _, f = range cachefiles {
		infos = f.GetPropOpt(id, prop)
		if infos != nil {
			return infos
		}
	}

	return nil
}

func GetEntityByConfig(id string) (string, error) {
	var f *csvparser
	var infos []string
	var fi string
	for fi, f = range cachefiles {
		if strings.HasPrefix(fi, "conf_") {
			continue
		}
		infos = f.Find(id)
		if infos != nil {
			break
		}
	}

	if infos == nil {
		return "", errors.New("config not found")
	}

	if f.entity == 0 {
		return "", errors.New("entity not define," + id)
	}

	typ := infos[f.entity-1]
	return typ, nil
}

func LoadStructByFile(file string, id string, dest interface{}) error {
	var f *csvparser
	var infos []string
	var ok bool
	if f, ok = cachefiles[file]; !ok {
		return errors.New("config not found")
	}

	infos = f.Find(id)
	if infos == nil {
		return errors.New("config not found")
	}

	dpv := reflect.ValueOf(dest)
	if dpv.Kind() != reflect.Ptr {
		return errors.New("dest not a pointer")
	}

	for k, hi := range f.heads {
		dfv := dpv.Elem().FieldByName(hi.name)
		if !dfv.IsValid() {
			continue
		}
		dv := reflect.Indirect(dfv)

		switch dv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if infos[k] == "" {
				continue
			}
			i64, err := strconv.ParseInt(infos[k], 10, dv.Type().Bits())
			if err != nil {
				return fmt.Errorf("converting string %q to a %s: %v", infos[k], dv.Kind(), err)
			}
			dv.SetInt(i64)
			continue
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if infos[k] == "" {
				continue
			}
			u64, err := strconv.ParseUint(infos[k], 10, dv.Type().Bits())
			if err != nil {
				return fmt.Errorf("converting string %q to a %s: %v", infos[k], dv.Kind(), err)
			}
			dv.SetUint(u64)
			continue
		case reflect.Float32, reflect.Float64:
			if infos[k] == "" {
				continue
			}
			f64, err := strconv.ParseFloat(infos[k], dv.Type().Bits())
			if err != nil {
				return fmt.Errorf("converting string %q to a %s: %v", infos[k], dv.Kind(), err)
			}
			dv.SetFloat(f64)
			continue
		case reflect.String:
			dv.SetString(infos[k])
			continue
		default:
			log.LogError(dv.Kind(), " not support")
			continue
		}
	}
	return nil
}

func LoadConfigByFile(file string, id string, ent PropSeter) error {
	var f *csvparser
	var infos []string
	var ok bool
	if f, ok = cachefiles[file]; !ok {
		return errors.New("config not found")
	}

	infos = f.Find(id)
	if infos == nil {
		return errors.New("config not found")
	}

	for k, hi := range f.heads {
		if hi.name == "Entity" || strings.HasSuffix(hi.name, "_script") {
			continue
		}
		err := SetProp(ent, hi.name, infos[k])
		if err != nil {
			log.LogWarning(err, ":", hi.name)
		}
	}
	return nil
}

func LoadFromConfig(id string, ent PropSeter) error {
	var f *csvparser
	var infos []string
	var fi string
	for fi, f = range cachefiles {
		if strings.HasPrefix(fi, "conf_") {
			continue
		}
		infos = f.Find(id)
		if infos != nil {
			break
		}
	}

	if infos == nil {
		return errors.New("config not found")
	}

	for k, hi := range f.heads {
		if hi.name == "Entity" || strings.HasSuffix(hi.name, "_script") {
			continue
		}
		err := SetProp(ent, hi.name, infos[k])
		if err != nil {
			log.LogWarning(err, ":", hi.name)
		}
	}

	return nil
}

func SetProp(ent PropSeter, name string, val string) error {
	typ, _, err := ent.PropertyType(name)
	if err != nil {
		return nil
	}

	switch typ {
	case DT_INT8, DT_INT16, DT_INT32, DT_INT64, DT_UINT8, DT_UINT16, DT_UINT32, DT_UINT64, DT_FLOAT32, DT_FLOAT64, DT_STRING:
		ent.Set(name, val)
		break
	default:
		errors.New("type not support")
		break
	}

	return nil
}
