#!/bin/bash

cd /opt/www/MyAPI/dev/current

export PORT=3004
export LOG_SQL=true
export DEV_MODE=true

./myapidev
