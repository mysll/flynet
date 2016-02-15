package main

import (
	"db"
	"flag"
	"fmt"
)

var (
	dbname = flag.String("db", "", "dbname")
	source = flag.String("ds", "", "datasource")
	path   = flag.String("c", "", "entity path")
	help   = flag.Bool("h", false, "help")
	drop   = flag.Bool("d", false, "Drop db, this option will drop db, be careful to use!!!!!!")
	role   = flag.String("r", "Player", "role entity")
	useage = "usage:dbmgr -db \"dbname\" -ds \"datasource\" -c \"eitity path\" [-r \"role entity name, default is Player\"] [-h] [-d] "
)

func main() {
	flag.Parse()
	if *help {
		fmt.Println(useage)
		flag.PrintDefaults()
		return
	}

	if *dbname == "" {
		fmt.Println(useage)
		flag.PrintDefaults()
		return
	}

	if *source == "" {
		fmt.Println(useage)
		flag.PrintDefaults()
		return
	}

	if *path == "" {
		fmt.Println(useage)
		flag.PrintDefaults()
		return
	}

	sync := db.CreateSyncDB(*dbname, *source)
	sync.SyncDB(*path, *drop, *role)

}
