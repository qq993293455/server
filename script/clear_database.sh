#!/bin/bash

ip=10.23.20.53
passwd=F3xS!UB8PYGs
user=root
port=3306

getCLearSql(){
  mysql -u$1 -p$2 -h$3 -P$4 -e "select CONCAT('truncate TABLE ',table_schema,'.',table_name, ';') as 'use $5;' from information_schema.TABLES where table_schema ='$5'" > $5.sql
}

doClearSql(){
  mysql -u$1 -p$2 -h$3 -P$4 -e "source $5.sql"
}

clearSql(){
  getCLearSql $1 $2 $3 $4 $5
  doClearSql $1 $2 $3 $4 $5
}

doClear(){
   for i in `mysql -u$1 -p$2 -h$3 -P$4 -e "select table_name as 'TTTTT' from information_schema.TABLES where table_schema ='$5'"|grep -v TTTTT`
   do
    mysql -u$1 -p$2 -h$3 -P$4 -e "truncate TABLE $5.$i;"
   done
}

doClear $user $passwd $ip $port im
doClear $user $passwd $ip $port rank
doClear $user $passwd $ip $port game

mysql -u$user -p$passwd -h$ip -P$port -e "INSERT INTO game.admin_user(username,password,role,status,created_at)VALUES('op','a10f7d863326aa6d8cff9b7e25a6db03',100,0,0);"
mysql -u$user -p$passwd -h$ip -P$port -e "INSERT INTO game.admin_user(username,password,role,status,created_at)VALUES('admin','68d0b24f44311236f4140d274e3ad88a',100,0,0);"