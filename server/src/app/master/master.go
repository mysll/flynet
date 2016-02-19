package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"libs/log"
	"master"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bitly/go-simplejson"
)

var (
	config = flag.String("c", "../config", "config path")
)

func readConfig(path string) (json *simplejson.Json, err error) {

	body, e := ioutil.ReadFile(path)
	if e != nil {
		err = e
		return
	}
	json, err = simplejson.NewJson(body)
	return
}

func main() {
	_, err := os.Stat("log")
	if err != nil {
		os.Mkdir("log", os.ModePerm)
	}

	if _, err := os.Stat("log/master.log"); err == nil {
		os.Remove("log/master.log")
	}

	log.WriteToFile("log/master.log")

	defer func() {
		if e := recover(); e != nil {
			log.LogFatalf(e)
			time.Sleep(1 * time.Second)
		}
	}()

	flag.Parse()
	if *config == "" {
		flag.PrintDefaults()
		return
	}

	m := master.NewMaster()
	apps := *config + "/app.json"
	appdata, e := ioutil.ReadFile(apps)
	if e != nil {
		panic(e)
	}

	var app master.App
	if err := json.Unmarshal(appdata, &app); err != nil {
		panic(err)
	}

	m.AppDef = app
	servers := *config + "/servers.json"
	json, err := readConfig(servers)
	if err != nil {
		panic(err)
	}

	def, err1 := json.Map()
	if err1 != nil {
		panic(err1)
	}
	for key := range def {
		if key == "master" {
			mst := json.Get(key)
			if host, ok := mst.CheckGet("host"); ok {
				v, err := host.String()
				if err != nil {
					panic(err)
				}

				m.Host = v
			} else {
				m.Host = "127.0.0.1"
			}

			if localip, ok := mst.CheckGet("localip"); ok {
				v, err := localip.String()
				if err != nil {
					panic(err)
				}

				m.LocalIP = v
			} else {
				m.LocalIP = "127.0.0.1"
			}

			if outerip, ok := mst.CheckGet("outerip"); ok {
				v, err := outerip.String()
				if err != nil {
					panic(err)
				}

				m.OuterIP = v
			} else {
				m.OuterIP = "127.0.0.1"
			}

			if port, ok := mst.CheckGet("port"); ok {
				v, err := port.Int()
				if err != nil {
					panic(err)
				}

				m.Port = v
			} else {
				m.Port = 5100
			}

			if agent, ok := mst.CheckGet("agent"); ok {
				v, err := agent.Bool()
				if err != nil {
					panic(err)
				}

				m.Agent = v
			}

			if agentid, ok := mst.CheckGet("agentid"); ok {
				v, err := agentid.String()
				if err != nil {
					panic(err)
				}

				m.AgentId = v
			}

			if cp, ok := mst.CheckGet("consoleport"); ok {
				v, err := cp.Int()
				if err != nil {
					panic(err)
				}

				m.ConsolePort = v
			}

			if tpl, ok := mst.CheckGet("consoleroot"); ok {
				v, err := tpl.String()
				if err != nil {
					panic(err)
				}
				log.LogMessage("path:", v)
				m.Template = v
			}

			if waits, ok := mst.CheckGet("waitagents"); ok {
				v, err := waits.Int()
				if err != nil {
					panic(err)
				}

				m.WaitAgents = v
			}

			if startuid, ok := mst.CheckGet("startuid"); ok {
				v, err := startuid.Int()
				if err != nil {
					panic(err)
				}

				master.AppUid = int32(v)
			}

			if nobalance, ok := mst.CheckGet("nobalance"); ok {
				v, err := nobalance.Bool()
				if err != nil {
					panic(err)
				}

				m.NoBalance = v
			}

			continue

		}

		m.AppArgs[key], _ = json.Get(key).MarshalJSON()
	}

	exitChan := make(chan int)
	signalChan := make(chan os.Signal, 1)
	go func() {
		<-signalChan
		exitChan <- 1
	}()
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	m.Start()
	m.Wait(exitChan)
	m.Exit()
	log.CloseLogger()
}
