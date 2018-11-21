#!/usr/bin/env bash
# @describe: 定时在线修改数据库结构
# @author:   Jerry Yang(hy0kle@gmail.com)

# 定时执行修改数据库字段,UTC-20:30(印尼04:30)为业务流量低峰
#30 20 * * * cd /home/nginx/auto-ost && ./online-schema-change.sh >> /work/micro-loan/logs/osc.log 2>&1

#set -x

timeStr=`date +"%Y-%m-%d %R:%S"`
dateStr=`date +"%Y-%m-%d"`

function timeNow() {
    timeStr=`date +"%Y-%m-%d %R:%S"`
}

dbapi='mysql -umicroloan -hrupiahcepatweb.cj6kgbqqvjzo.ap-southeast-1.rds.amazonaws.com microloan -pnGrMqVxB3Q'
dbadmin='mysql --default-character-set=utf8mb4 -umicroloan_admin -hrupiahcepatweb.cj6kgbqqvjzo.ap-southeast-1.rds.amazonaws.com microloan_admin -pp9GmExcLmy'

apiSQL='api.sql'
adminSQL='admin.sql'

# 修改api表结构
if [ ! -f "$apiSQL" ]
then
    echo "[$timeStr] can NOT find file: $apiSQL"
else
    echo "[$timeStr] exec sql: $apiSQL"
    $dbapi < $apiSQL
    timeNow
    mv "$apiSQL" "$apiSQL.$dateStr"
    # TODO: 出错应该报警
fi

#sleep 5

# 修改admin表结构
if [ ! -f "$adminSQL" ]
then
    echo "[$timeStr] can NOT find file: $adminSQL"
else
    echo "[$timeStr] exec sql: $adminSQL"
    $dbadmin < $adminSQL
    timeNow
    mv "$adminSQL" "$adminSQL.$dateStr"
    # TODO: 出错应该报警
fi

timeNow
echo "[$timeStr] jobs done"

# vim:set ts=4 sw=4 et fdm=marker:
