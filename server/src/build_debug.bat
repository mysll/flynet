del ..\pkg\* /s/q
go build -gcflags "-N -l" -o ../../bin/tools/data.exe ./app/tools/data/
go build -gcflags "-N -l" -o ../../bin/tools/dbsync.exe ./app/tools/dbsync/
go build -gcflags "-N -l" -o ../../bin/tools/rpcmaker ./app/tools/rpcmaker/
go build -gcflags "-N -l" -o ../../bin/master.exe ./app/master/
go build -gcflags "-N -l" -o ../../bin/db.exe ./app/db/
go build -gcflags "-N -l" -o ../../bin/baseapp.exe ./app/base/
go build -gcflags "-N -l" -o ../../bin/basemgr.exe ./app/basemgr/
go build -gcflags "-N -l" -o ../../bin/loginapp.exe ./app/login/
go build -gcflags "-N -l" -o ../../bin/loginmgr.exe ./app/loginmgr/