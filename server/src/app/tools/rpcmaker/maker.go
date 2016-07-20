package main

import (
	"bufio"
	"encoding/xml"
	"io/ioutil"
	"os"
	"server/libs/log"
	"sort"
	"strings"
	"text/template"
)

var (
	tpl = `<rpc>{{range .Infos}}
	<call>
		<node>{{.RpcNode}}</node>
		<service>{{.RpcService}}</service>
		<method>{{.RpcMethod}}</method>
		<id>{{.Id}}</id>
	</call>{{end}}
</rpc>`
)

type Call struct {
	Service string `xml:"service"`
	Method  string `xml:"method"`
	Id      int16  `xml:"id"`
}

type Rpc struct {
	Calls []Call `xml:"call"`
}

type RpcInfo struct {
	RpcNode    string
	RpcService string
	RpcMethod  string
	Id         int16
}

type RpcCollection struct {
	Infos []RpcInfo
}

func (rc RpcCollection) Len() int {
	return len(rc.Infos)
}

func (rc RpcCollection) Less(i, j int) bool {
	if rc.Infos[i].RpcNode != rc.Infos[j].RpcNode {
		return rc.Infos[i].RpcNode < rc.Infos[j].RpcNode
	}
	return rc.Infos[i].Id < rc.Infos[j].Id
}

func (rc RpcCollection) Swap(i, j int) {
	rc.Infos[i], rc.Infos[j] = rc.Infos[j], rc.Infos[i]
}

var (
	path = "../interface"
)

func main() {
	dir, _ := os.Open(path)
	files, _ := dir.Readdir(0)
	var output RpcCollection
	exists := map[string]bool{}
	id := 0
	for _, file := range files {
		if !file.IsDir() {
			if file.Name() == "rpc.xml" {
				continue
			}
			filename := strings.Replace(file.Name(), "_rpc.xml", "", 0)
			vals := strings.Split(filename, "_")
			node := vals[0]
			if _, ok := exists[node]; ok {
				continue
			}
			id++
			exists[node] = true
			log.LogMessage(node, ",", file.Name())
			f, err := os.Open(path + "/" + file.Name())
			if err != nil {
				panic(err)
			}

			defer f.Close()
			data, err := ioutil.ReadAll(f)
			if err != nil {
				panic(err)
			}
			var obj Rpc

			err = xml.Unmarshal(data, &obj)
			if err != nil {
				panic(err.Error() + file.Name())
			}

			log.LogMessage(obj.Calls)
			for _, c := range obj.Calls {
				output.Infos = append(output.Infos, RpcInfo{node, c.Service, c.Method, c.Id + int16(id*1000)})
			}

		}
	}

	sort.Sort(output)
	t, err := template.New("maker").Parse(tpl)
	if err != nil {
		log.LogError(err.Error())
	}

	if err != nil {
		log.LogError("template", err)
	}

	//save file
	file, err := os.Create(path + "/rpc.xml")
	if err != nil {
		log.LogError("writer", err)
		return
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	err = t.Execute(writer, output)
	if err != nil {
		log.LogError("writer", err)
	}

	writer.Flush()

}
