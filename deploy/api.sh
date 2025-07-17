#!/bin/bash

scp scripts/deploy.sh "myusername@$1:/opt/www/MyAPI/scripts/"
scp scripts/startProduction.sh "damian@$1:/opt/www/MyAPI/api/shared/"
scp scripts/wait-for-it.sh "damian@$1:/opt/www/MyAPI/api/shared/"

echo "deploying API to $1"
ssh "damian@$1" bash /opt/www/MyAPI/scripts/deploy.sh
