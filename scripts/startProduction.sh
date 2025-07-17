#!/bin/bash

cd /opt/www/MyAPI/api/current

export PORT=3003
export LOG_SQL=true
 

#/bin/sh -c bin/scripts/wait-for-it.sh 10.0.0.1:5432 -- 
./myapi
