package base

import (
	"data/datatype"
	"data/helper"
	"libs/log"
)

type KeyVal struct {
	UniqueName string
	Value      string
}

type PropRatio struct {
	ID      string
	BTRatio string
}

type Config struct {
	configMap map[string]*KeyVal
}

func NewConfig() *Config {
	conf := &Config{}
	return conf
}

func (this *Config) GetInt64(key string) int64 {
	var ret int64
	if val, ok := this.configMap[key]; ok {
		datatype.ParseStrNumber(val.Value, &ret)
		return ret
	}

	log.LogError("config key ", key, " not found")
	return ret
}

func (this *Config) GetUint64(key string) uint64 {
	var ret uint64
	if val, ok := this.configMap[key]; ok {
		datatype.ParseStrNumber(val.Value, &ret)
		return ret
	}

	log.LogError("config key ", key, " not found")
	return ret
}

func (this *Config) GetFloat32(key string) float32 {
	var ret float32
	if val, ok := this.configMap[key]; ok {
		datatype.ParseStrNumber(val.Value, &ret)
		return ret
	}

	log.LogError("config key ", key, " not found")
	return ret
}

func (this *Config) GetFloat64(key string) float64 {
	var ret float64
	if val, ok := this.configMap[key]; ok {
		datatype.ParseStrNumber(val.Value, &ret)
		return ret
	}

	log.LogError("config key ", key, " not found")
	return ret
}

func (this *Config) GetString(key string) string {
	if val, ok := this.configMap[key]; ok {
		return val.Value
	}

	log.LogError("config key ", key, " not found")
	return ""
}

func (this *Config) load() {
	ids := helper.GetConfigIds("conf_configuration.csv")

	if len(ids) == 0 {
		log.LogWarning("conf_configuration.csv file load failed")
		return
	}

	this.configMap = make(map[string]*KeyVal, len(ids))

	for _, id := range ids {
		conf := &KeyVal{}

		if err := helper.LoadStructByFile("conf_configuration.csv", id, conf); err != nil {
			log.LogFatalf(err)
			continue
		}

		this.configMap[conf.UniqueName] = conf
	}

	log.LogMessage("conf_configuration.csv file load ok")

	ids = helper.GetConfigIds("conf_bp_ratio.csv")

	if len(ids) == 0 {
		log.LogWarning("conf_bp_ratio.csv file load failed")
		return
	}

	for _, id := range ids {
		conf := &KeyVal{}
		tmp := &PropRatio{}
		if err := helper.LoadStructByFile("conf_bp_ratio.csv", id, tmp); err != nil {
			log.LogFatalf(err)
			continue
		}

		conf.UniqueName = "bp_" + tmp.ID
		conf.Value = tmp.BTRatio
		this.configMap[conf.UniqueName] = conf
	}

	log.LogMessage("conf_bp_ratio.csv file load ok")

}

func (this *Config) GetConfigByKey(key string) string {
	if val, ok := this.configMap[key]; ok {
		return val.Value
	}

	return "nil"
}
