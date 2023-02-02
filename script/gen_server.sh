#!/bin/bash

cd coin-server
branch=$(git rev-parse --abbrev-ref HEAD)
cd ..
cd out
if [[ $? -ne 0 ]]; then
  mkdir "out"
  cd out
fi

time=$(date "+%Y%m%d%H%M%S")
dir="$branch"-"$time"
mkdir $dir

cd ../coin-server
cd dungeon-match-server
go build -o dungeon-match-server.exe
mv dungeon-match-server.exe ../../out/$dir

cd ../game-server
go build -o game-server.exe
mv game-server.exe ../../out/$dir

cd ../gateway-tcp
go build -o gateway.exe
mv gateway.exe ../../out/$dir

cd ../gen-rank-server
go build -o gen-rank-server.exe
mv gen-rank-server.exe ../../out/$dir

cd ../guild-filter-server
go build -o guild-filter-server.exe
mv guild-filter-server.exe ../../out/$dir

cd ../new-battle-server
go build -o new-battle-server.exe
mv new-battle-server.exe ../../out/$dir

cd ../recommend-server
go build -o recommend-server.exe
mv recommend-server.exe ../../out/$dir
