#!/bin/bash

# 根据输入参数执行相应的操作
case "$1" in
    "start")
        echo "Starting Service..."
        nohup ./bt ../config/config.json > ./logs/nohup-bt.log 2>&1 &
        ;;
    "stop")
        echo "Stopping Service..."
        pkill -f "./bt ../config/config.json"
        ;;
    *)
        echo "Usage: $0 {start|stop}"
        exit 1
        ;;
esac