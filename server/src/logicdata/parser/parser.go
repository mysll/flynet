package parser

import (
	"encoding/xml"
	"io/ioutil"
	"os"
	"server/libs/log"
)

type Property struct {
	Name      string `xml:"name,attr"`
	Tag       string `xml:"tag,attr"`
	Type      string `xml:"type,attr"`
	Len       int    `xml:"len,attr"`
	Save      string `xml:"save,attr"`
	Public    string `xml:"public,attr"`
	Comment   string `xml:"comment,attr"`
	SceneData string `xml:"scenedata,attr"`
	Realtime  string `xml:"realtime,attr"`
}

type Column struct {
	Type    string `xml:"type,attr"`
	Name    string `xml:"name,attr"`
	Tag     string `xml:"tag,attr"`
	Comment string `xml:"comment,attr"`
	Len     int    `xml:"len,attr"`
}

type Record struct {
	Name      string   `xml:"name,attr"`
	Cols      int      `xml:"cols,attr"`
	Maxrows   int      `xml:"maxrows,attr"`
	Save      string   `xml:"save,attr"`
	Visible   string   `xml:"visible,attr"`
	Comment   string   `xml:"comment,attr"`
	Columns   []Column `xml:"column"`
	SceneData string   `xml:"scenedata,attr"`
	Type      string   `xml:"type,attr"`
}

type Object struct {
	Name       string     `xml:"name"`
	Type       string     `xml:"type"`
	Base       string     `xml:"base"`
	Persistent string     `xml:"persistent"`
	Include    string     `xml:"include"`
	Interfaces []string   `xml:"interfaces>interface"`
	Propertys  []Property `xml:"propertys>property"`
	Records    []Record   `xml:"records>record"`
	flag       bool       //include 处理标志
}

var (
	Defs = make(map[string]*Object)
)

func parseEntity(file string) *Object {
	f, err := os.Open(file)
	if err != nil {
		panic(err)
	}

	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}
	obj := &Object{}
	err = xml.Unmarshal(data, obj)
	if err != nil {
		panic(err.Error() + file)
	}

	return obj
}

func parseInclude(obj *Object) {
	if obj.Include != "" {
		if obj.Include == obj.Name {
			log.LogFatalf(obj.Name, " include self")
		}

		parent := Defs[obj.Include]
		if parent == nil {
			log.LogFatalf(obj.Name, " include ", obj.Include, " not found")
		}

		if !parent.flag {
			parseInclude(parent)
		}

		obj.Propertys = append(parent.Propertys, obj.Propertys...)
		obj.Records = append(parent.Records, obj.Records...)
	}
	obj.flag = true
}

func LoadAllDef(path string) {
	dir, _ := os.Open(path)
	files, _ := dir.Readdir(0)

	defs := make(map[string]*Object, 20)
	for _, f := range files {
		if !f.IsDir() {
			obj := parseEntity(path + "/" + f.Name())
			defs[obj.Name] = obj
		}
	}

	Defs = defs

	for _, v := range Defs {
		parseInclude(v)
	}

}
