#!/bin/sh

 sleep 1
 echo "nc -z $LOCUST_MASTER_HOST $LOCUST_MASTER_PORT"
 while ! nc -z $LOCUST_MASTER_HOST $LOCUST_MASTER_PORT
 do
   echo waiting for locust ready ...
   sleep 2
 done

echo locust is ready,starting load-test ...
./load-test
