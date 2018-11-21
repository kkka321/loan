#!/usr/bin/env bash
# @describe: 重启admin
# @author:   Jerry Yang(hy0kle@gmail.com)

#set -x

function time_now() {
    timeNow=`date +"%Y-%m-%d %R:%S"`
}

time_now
echo "[$timeNow] ---Before---"
curl -XPOST http://127.0.0.1:8600/ping
echo ""

ps aux | grep admin:8600 | grep -v grep
ps aux | grep admin:8600 | grep -v grep | awk '{print " kill -9 "$2}' | sh

sleep 3

time_now
echo "[$timeNow] ---After---"
ps aux | grep admin:8600 | grep -v grep
curl -XPOST http://127.0.0.1:8600/ping
echo ""
echo "[$timeNow] reboot server admin done"
# vim:set ts=4 sw=4 et fdm=marker:

