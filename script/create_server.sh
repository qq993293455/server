#!/bin/bash

path=$(pwd)
serverExist=$(ls -a | grep coin-server)
clientExist=$(ls -a | grep l5client)
shareExist=$(ls -a | grep share)

createBranch(){
  cd coin-server
  git rev-parse --verify $1
  if [[ $? -ne 0 ]]; then
    echo "-----------------------------开始新建分支-----------------------------"
    git stash
    git checkout develop
    git clean -df
    git reset --hard
    git pull --rebase
    git branch $1
    git checkout $1
    git push --set-upstream origin $1
    git pull --rebase
    git push

    cd ../share
    git stash
    git checkout develop
    git clean -df
    git reset --hard
    git pull --rebase
    git branch $1
    git checkout $1
    git push --set-upstream origin $1
    git pull --rebase
    git push

    cd ../l5client
    git stash
    git checkout develop
    git clean -df
    git reset --hard
    git pull --rebase
    git branch $1
    git checkout $1
    git push --set-upstream origin $1
    git pull --rebase
    git push
    echo "-----------------------------新建分支完成-----------------------------"
    exit 0
  fi

  cd ..
  updateProject $1
  exit 0
}

updateProject(){
  echo "-----------------------------开始更新工程-----------------------------"
  cd coin-server
  git clean -df
  git reset --hard
  git pull --rebase
  git checkout $1
  git clean -df
  git reset --hard
  git pull --rebase
  cd ../l5client
  git clean -df
  git reset --hard
  git pull --rebase
  git checkout $1
  git clean -df
  git reset --hard
  git pull --rebase
  cd ../share
  git clean -df
  git reset --hard
  git pull --rebase
  git checkout $1
  git clean -df
  git reset --hard
  git pull --rebase
  echo "-----------------------------更新完成-----------------------------"
}


if [[ -n "$serverExist" ]] && [[ -n "$shareExist" ]] && [[ -n "$clientExist" ]]; then
  if [ $# == 1 ] ; then
    echo $1
    createBranch $1
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
  createBranch $1
  exit 0
fi

