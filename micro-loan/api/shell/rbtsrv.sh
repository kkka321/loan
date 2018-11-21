#!/usr/bin/env bash
# @describe: 重启api服务
# @author:   Jerry Yang(hy0kle@gmail.com)

#set -x

function time_now() {
    timeNow=`date +"%Y-%m-%d %R:%S"`
}

time_now
echo "[$timeNow] ---Before---"
curl -XPOST http://127.0.0.1:8700/ping
echo ""

ps aux | grep microloan-api:8700 | grep -v grep
ps aux | grep microloan-api:8700 | grep -v grep | awk '{print " kill -9 "$2}' | sh

sleep 3

time_now
echo "[$timeNow] ---After---"
ps aux | grep microloan-api:8700 | grep -v grep
curl -XPOST http://127.0.0.1:8700/ping
echo ""
echo "[$timeNow] reboot microloan-api:8700  done"
# vim:set ts=4 sw=4 et fdm=marker:

