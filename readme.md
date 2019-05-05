[![Build Status](https://travis-ci.com/codeuniversity/al-master.svg?branch=master)](https://travis-ci.com/codeuniversity/al-master)

##Setup:

0) Download ```two-masters``` branch from ```al-master```, ```al-cis``` & ```al-proto```
1) start first master with ```make run```
2) start second master with ```go run main/main.go -grpc_port 3001 -http_port 4001 -chief_master_address localhost:3000```
3) start CIS with ```make dep``` then ```make run```
