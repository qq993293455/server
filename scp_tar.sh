#!/usr/bin/expect
set timeout 600
set tarfile [lindex $argv 0]  
set passwd RfyzN2lp45pX\\Ro
spawn scp $tarfile igguser@46.137.250.44:/home/igguser/gocd/bin/
expect "*password:"
send "$passwd\r"
expect eof
