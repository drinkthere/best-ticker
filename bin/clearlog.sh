#!/bin/bash

# 指定日志文件所在的目录
aws_log_directory="/home/ubuntu/code/best-ticker/bin/logs"
aliyun_log_directory="/root/code/best-ticker/bin/logs"
# 保留最近的两个okxmm.log文件
ls -t $aws_log_directory/* | tail -n +5 | xargs rm -f
ls -t $aliyun_log_directory/* | tail -n +5 | xargs rm -f
