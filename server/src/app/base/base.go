package main

import (
	. "base"
	"flag"
	"github.com/bitly/go-simplejson"
	"runtime"
	"server"
)

var (
	master    = flag.String("m", "", "master info")
	appid     = flag.Int("d", 0, "appid")
	typ       = flag.String("t", "", "app type")
	startargs = flag.String("s", "", "start args")
)

func main() {
	flag.Parse()
	if *master == "" || *startargs == "" {
		flag.PrintDefaults()
		panic("args error")
	}

	json, err := simplejson.NewJson([]byte(*startargs))
	if err != nil {
		panic(err)
	}

	App.Server = server.NewServer(App, int32(*appid))
	runtime.GOMAXPROCS(runtime.NumCPU())
	if App.Start(*master, *typ, json) {
		App.Wait()
	}
}
