del ..\pkg\* /s/q
go build -o ../../bin/tools/data.exe ./app/tools/data/
go build -o ../../bin/tools/dbsync.exe ./app/tools/dbsync/
go build -o ../../bin/tools/rpcmaker.exe ./app/tools/rpcmaker/
go build -o ../../bin/master.exe ./app/master/
go build -o ../../bin/db.exe ./app/db/
go build -o ../../bin/baseapp.exe ./app/base/
go build -o ../../bin/basemgr.exe ./app/basemgr/
go build -o ../../bin/loginapp.exe ./app/login/
go build -o ../../bin/loginmgr.exe ./app/loginmgr/
go build -o ../../bin/statusapp.exe ./app/status/