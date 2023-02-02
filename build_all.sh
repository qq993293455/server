#!/bin/bash

set -e
set -x

branch=`git symbolic-ref --short -q HEAD`
dStr=`date "+%Y-%m-%d %H:%M:%S"`
dStr=`echo $dStr|sed 's/\ /_/g'|sed 's/:/_/g'`
goDir=golang-${branch}-${dStr}
mkdir -p ${goDir}
rm -rf ${goDir}/*

echo "go build start ..."
#git pull --rebase origin $BRANCH
go build  -o ./${goDir}/gateway ./gateway-tcp/main.go
go build  -o ./${goDir}/activity-server ./activity-server/main.go
go build  -o ./${goDir}/gameserver ./game-server/main.go
go build  -o ./${goDir}/pika-viewer ./pikaviewer/main.go
go build  -o ./${goDir}/roguelikematchserver ./roguelike-match-server/main.go
go build  -o ./${goDir}/stateserver ./role-state-server/main.go
go build  -o ./${goDir}/guild-filter ./guild-filter-server/main.go
go build  -o ./${goDir}/rankserver ./rank-server/main.go
go build  -o ./${goDir}/racingrankserver ./racingrank-server/main.go
go build  -o ./${goDir}/centerserver ./new-center-server/main.go
go build  -o ./${goDir}/edgeserver ./edge-server/main.go
go build  -o ./${goDir}/arenaserver ./arena-server/main.go
go build  -o ./${goDir}/syncrole ./sync-role-worker/main.go
go build  -o ./${goDir}/statistical-server ./statistical-server/main.go
go build  -o ./${goDir}/activity-ranking-server ./activity-ranking-server/main.go
go build  -o ./${goDir}/payserver ./pay-server/main.go
go build  -o ./${goDir}/noticeserver ./notice-server/main.go
go build  -o ./${goDir}/fashionserver ./fashion-server/main.go

echo "go build done"

tar -zcvf ${goDir}.tar.gz ./${goDir}

#expect scp_tar.sh ${goDir}.tar.gz
rm -rf ${goDir}
