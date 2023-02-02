#!/bin/bash

if [ $# != 1 ] ; then
  echo "error, 请传入分支名称"
  exit 1
fi

cd out
dirs=$(ls -a | grep -E $1)
dir=""
for e in ${dirs[*]}
do
    if [[ $e == "$1" ]]; then
        dir=$1
        break
    fi
done

if [[ $dir == "" ]]; then
  echo "未找到对应文件夹"
  exit 1
fi

branch=$(echo $dir | awk -F'-' '{print $1}')
echo $branch

cd $dir
echo "----------------------------开始启动服务器----------------------------"
export SERVER_ID=120
export RULE_TAG=$branch
nohup ./game-server.exe 2>&1 > game-server.log &
nohup ./gateway.exe 2>&1 > gateway.log &
nohup ./dungeon-match-server.exe 2>&1 > dungeon-match-server.log &
nohup ./gen-rank-server.exe 2>&1 > gen-rank-server.log &
nohup ./guild-filter-server.exe 2>&1 > guild-filter-server.log &
nohup ./new-battle-server.exe 2>&1 > new-battle-server.log &
nohup ./recommend-server.exe 2>&1 > recommend-server.log &
