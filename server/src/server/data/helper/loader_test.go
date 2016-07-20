package helper

import (
	"testing"
)

type TestLevelCost struct {
	SilverCost_QiangHua   int32
	SkillCostExp_Qianghua int32
	PetCostExp_QiangHua   int32
}

func TestLoadStruct(t *testing.T) {
	LoadAllConfig("C:/home/work/goserver/assets")
	cost := &TestLevelCost{}
	LoadStructByFile("conf_level_cost.csv", "19", cost)
	t.Log(cost)
}
