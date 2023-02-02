#!/bin/bash

path=$(pwd)
serverExist=$(ls -a | grep coin-server)
clientExist=$(ls -a | grep l5client)
shareExist=$(ls -a | grep share)

getBranch(){
  cd coin-server
  git rev-parse --verify $1
  if [[ $? -ne 0 ]]; then
    echo "----------------------------找不到分支------------------------------"
  fi
  cd ..
  updateProject $1
  exit 0
}

updateProject(){
  echo "-----------------------------开始更新工程-----------------------------"
  cd coin-server
  git stash
  git clean -df
  git reset --hard
  git pull --rebase
  git checkout $1
  git fetch origin $1
  git pull --rebase
  cd ../l5client
  git stash
  git clean -df
  git reset --hard
  git pull --rebase
  git checkout $1
  git fetch origin $1
  git pull --rebase
  cd ../share
  git stash
  git clean -df
  git reset --hard
  git pull --rebase
  git checkout $1
  git fetch origin $1
  git pull --rebase
  echo "-----------------------------更新完成-----------------------------"
}


if [[ -n "$serverExist" ]] && [[ -n "$shareExist" ]] && [[ -n "$clientExist" ]]; then
  if [ $# == 1 ] ; then
    echo $1
    getBranch $1
    exit 0
  fi
  updateProject $(git rev-parse --abbrev-ref HEAD)
  exit 0
fi

echo "-----------------------------开始拉取工程-----------------------------"
git clone git@gitlab.cdl5.org:chengdu-l5/coin-server.git
git clone git@gitlab.cdl5.org:chengdu-l5/share.git
git clone git@gitlab.cdl5.org:chengdu-l5/l5client.git

if [ $# == 1 ] ; then
  echo $1
  getBranch $1
  exit 0
fi

