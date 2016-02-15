#!/bin/sh

rm -rf ../pkg/*
go build -o ../../bin/tools/data ./app/tools/data/
go build -o ../../bin/tools/dbsync ./app/tools/dbsync/
go build -o ../../bin/tools/rpcmaker ./app/tools/rpcmaker/
go build -o ../../bin/master ./app/master/
go build -o ../../bin/db ./app/db/
go build -o ../../bin/baseapp ./app/base/
go build -o ../../bin/basemgr ./app/basemgr/
go build -o ../../bin/loginapp ./app/login/
go build -o ../../bin/loginmgr ./app/loginmgr/
go build -o ../../bin/statusapp ./app/status/
