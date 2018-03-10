#!/bin/sh

# default timezone
if [ ! -n "$TZ" ]; then
    export TZ="Asia/Shanghai"
fi

# set timezone
ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && \
echo $TZ > /etc/timezone 

# k8s config  switch
if [ -f "/usr/local/zhiya/config/.zhiya.yaml" ]; then
    ln -s  /usr/local/zhiya/config/.zhiya.yaml /usr/local/zhiya/.zhiya.yaml
fi

# apply config
echo "===start==="
cat /usr/local/zhiya/.zhiya.yaml
echo "====end===="

# run command
if [ ! -z "$1" ]; then
    /usr/local/zhiya/zhiya $@
fi
