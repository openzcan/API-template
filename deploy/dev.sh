#!/bin/bash

scp scripts/deploydev.sh "damian@$1:/opt/www/MyAPI/scripts/"
scp scripts/startDev.sh "damian@$1:/opt/www/MyAPI/dev/shared/" 


echo "deploying DEV to $1"
ssh "damian@$1" bash /opt/www/MyAPI/scripts/deploydev.sh
 