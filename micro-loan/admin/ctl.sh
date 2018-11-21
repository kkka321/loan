#!/usr/bin/env bash
# @describe: 任务控制脚本,简化操作
# @author:   Jerry Yang(hy0kle@gmail.com)

#set -x

function usage() {
    echo "$0 cmd task"
    echo "    cmd:  stop|start|restart|status"
    echo "    task: identity_detect|need_review_order|wait4loan_order|event_push|timer_task|ticket_realtime_assign_task|monitor|schema|info_review_auto_call_task|all"
    exit 0
}

if [ $# -lt 2 ]
then
    usage
fi
cmd=$1
task=$2
if [ "$cmd" != "stop" ] && [ "$cmd" != "start" ] && [ "$cmd" != "restart" ] && [ "$cmd" != "status" ] && [ "$cmd" != "clear" ]
then
    usage
fi
if [ "$task" != "identity_detect" ] && [ "$task" != "need_review_order" ] && [ "$task" != "wait4loan_order" ] && [ "$task" != "event_push" ] && [ "$task" != "timer_task" ] && [ "$task" != "ticket_realtime_assign_task" ] && [ "$task" != "monitor" ] && [ "$task" != "schema" ] && [ "$task" != "info_review_auto_call_task" ] && [ "$task" != "all" ]
then
    usage
fi

WORKSPACE=$(cd $(dirname $0)/; pwd)
cd $WORKSPACE

logs_path="/work/micro-loan/logs/admin"
redis_host="microl.fhdzvb.ng.0001.apse1.cache.amazonaws.com"
redis_client="/usr/local/bin/redis-cli"

# 开发环境
if [ "x$TASK_ENV" == "xdev" ]
then
    logs_path="$WORKSPACE/logs"
    redis_host="127.0.0.1"
fi

if [ ! -d "$logs_path" ]
then
    echo "请指定工作环境: TASK_ENV"
    usage
fi

# 工作进程的名字
proc_title="micro-loan-cli-task:$task"

function getPid() {
    pid=`ps aux | grep $proc_title | grep -v grep | awk '{print $2}'`
}

# 当进程不存在时,退出
function checkAlive() {
    getPid
    if [ -z "$pid" ]
    then
        echo "$proc_title 工作进程不存在"
    else
        echo "$proc_title 正在工作中, pid: $pid"
    fi
}

function start() {
    getPid

    if [ -z "$pid" ]
    then
        echo "start $proc_title"
        #nohup ./task -name="$task" > "$logs_path/nohup.$task.log" &
        # 因为业务已经有日志,nohup的日志可忽略掉
        nohup ./task -name="$task" > /dev/null 2>&1 &
    else
        echo "$proc_title is working. pid: $pid"
    fi
}

function stop() {
    getPid

    if [ -z "$pid" ]
    then
        echo "$proc_title does not working, can't stop it."
    else
        $redis_client -h "$redis_host" RPUSH "queue:$task" "-111"
    fi

    sleep 1
    checkAlive
}

function restart() {
    stop
    sleep 1
    start
    sleep 1
    checkAlive
}

function doWork() {
    if [ "$cmd" == "start" ]
    then
        start
    fi

    if [ "$cmd" == "stop" ]
    then
        stop
    fi

    if [ "$cmd" == "restart" ]
    then
        restart
    fi

    if [ "$cmd" == "status" ]
    then
        checkAlive
    fi
}

if [ "$task" == "all" ]
then
    task_lists="identity_detect need_review_order wait4loan_order event_push timer_task ticket_realtime_assign_task monitor info_review_auto_call_task"
    for t in $task_lists
    do
        # 注意: 此处覆盖了全局变量!!!
        task="$t"
        proc_title="micro-loan-cli-task:$task"
        doWork
    done
else
    doWork
fi

# vim:set ts=4 sw=4 et fdm=marker:
