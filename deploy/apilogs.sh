#!/bin/bash

ssh root@$1 journalctl --vacuum-time=2d
ssh root@$1 journalctl "_SYSTEMD_UNIT=$2.service" > logs/$1-$2.txt
scp root@$1:/var/log/nginx/access.log logs/nginx-$1.log