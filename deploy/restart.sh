#!/bin/bash

echo "restart $1"
ssh root@$1 systemctl restart $2 
ssh root@$1 systemctl status $2
sleep $3