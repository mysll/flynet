package server

import (
	"bufio"
	"os"
	"server/libs/log"
	"server/libs/rpc"
	"text/template"
)

var (
	remotes  = make(map[string]interface{})
	handlers = make(map[string]interface{})
)

func GetRemote(name string) interface{} {
	if k, ok := remotes[name]; ok {
		return k
	}

	return nil
}

func GetHandler(name string) interface{} {
	if k, ok := handlers[name]; ok {
		return k
	}

	return nil
}

func GetAllHandler() map[string]interface{} {
	return handlers
}

func RegisterRemote(name string, remote interface{}) {
	if remote == nil {
		log.LogFatalf("rpc: Register remote is nil")
	}
	if _, dup := remotes[name]; dup {
		log.LogFatalf("rpc: Register called twice for remote " + name)
	}
	remotes[name] = remote
}

func RegisterHandler(name string, handler interface{}) {
	if handler == nil {
		log.LogFatalf("rpc: Register handler is nil")
	}
	if _, dup := handlers[name]; dup {
		log.LogFatalf("rpc: Register called twice for handler " + name)
	}
	handlers[name] = handler
}

var (
	tpl = `<rpc>{{range .Infos}}
	<call>
		<service>{{.RpcService}}</service>
		<method>{{.RpcMethod}}</method>
		<id>{{.Id}}</id>
	</call>{{end}}
</rpc>`
)

type RpcInfo struct {
	RpcService string
	RpcMethod  string
	Id         int16
}

type RpcCollection struct {
	Infos []RpcInfo
}

func createRpc(ch chan *rpc.RpcCall) *rpc.Server {
	rpc, err := rpc.CreateRpcService(remotes, handlers, ch)
	if err != nil {
		log.LogFatalf(err)
	}
	id := 0
	var collection RpcCollection
	for service, _ := range handlers {
		if service != "C2SHelper" {
			id++
			sid := id * 100
			info := rpc.GetRpcInfo("C2S" + service)
			for _, m := range info {
				sid++
				rinfo := RpcInfo{service, m, int16(sid)}
				collection.Infos = append(collection.Infos, rinfo)
			}
		}
	}
	if len(collection.Infos) > 0 {
		t, err := template.New("maker").Parse(tpl)
		if err != nil {
			log.LogError(err.Error())
		}

		if err != nil {
			log.LogError("template", err)
		}

		//save file
		file, err := os.Create("interface/" + core.Name + "_rpc.xml")
		if err != nil {
			log.LogError("writer", err)
			return rpc
		}
		defer file.Close()

		writer := bufio.NewWriter(file)

		err = t.Execute(writer, collection)
		if err != nil {
			log.LogError("writer", err)
		}

		writer.Flush()
	}
	return rpc
}
