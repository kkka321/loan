package main

import (
	"flag"
	"fmt"
	"os"
	//"time"

	// 数据库初始化
	_ "micro-loan/common/lib/db/mysql"
	"micro-loan/common/pkg/task"

	"github.com/astaxie/beego/logs"
	"github.com/erikdubbelboer/gspt"

	"micro-loan/common/lib/clogs"
	"micro-loan/common/tools"
	"micro-loan/common/types"

	"github.com/astaxie/beego"
)

/**
一期跑批任务限定单机,单进程,多协程,使用redis队列做解耦合
1. 生成队列
2. 消费队列
3. 安全退出
*/

const (
	programName = "micro-loan-cli-task"
)

var taskName string
var help bool
var version bool

func init() {
	flag.StringVar(&taskName, "name", "", "crontab, cli or backend `task-name`, need assign.")
	flag.BoolVar(&help, "h", false, "show usage and exit")
	flag.BoolVar(&version, "v", false, "show version and exit")

	// 改变默认的 Usage
	flag.Usage = usage
}

func usage() {
	fmt.Fprintf(os.Stderr, programName+` version: `+programName+`/`+types.TaskVersion+"\n"+
		"git-head-hash: "+tools.GitRevParseHead()+
		`
Usage: task [-hv] --name=[identity_detect|order_follow]

    identity_detect:                调用faceid进行身份证件,手持证件照比对识别
    need_review_order:              处理待审核订单
    wait4loan_order:                处理等待放款订单
    invalid_order:                  处理无效订单
    overdue_order:                  处理逾期订单
    event_push:                     处理事件
    timer_task:                     处理定时器任务
    monitor:                        系统监控
    repay_voice_order:              还款语音外呼
    overdue_auto_call:              逾期自动语音外呼
    ticket_realtime_assign_task:    处理需要被实时分配的工单
    auto_reduce_order:              自动减免跑批,满足条件的订单直接自动减免
    bigdata_contact:                大数据通讯录导入
    customer_recall:                召回用户
    register_remind:                注册提醒
    info_review_auto_call_task:     风控状态为‘等待自动外呼’的订单（InfoReview工单）自动外呼
    author_status_check:            授权状态检查

Options:
`)
	flag.PrintDefaults()
	os.Exit(0)
}

func showVersion() {
	fmt.Fprintf(os.Stderr, programName+` version: `+programName+`/`+types.TaskVersion+"\n")
	fmt.Fprintf(os.Stderr, "git-head-hash: %s\n", tools.GitRevParseHead())
	os.Exit(0)
}

func main() {
	flag.Parse()

	if help {
		flag.Usage()
	} else if version {
		showVersion()
	}

	// TODO: 如果需要通过命令行传参,此外的逻辑需要升级
	taskWork0Map := task.TaskWork0Map()
	// fmt.Print(taskWork0Map)
	// os.Exit(0)

	if _, ok := taskWork0Map[taskName]; !ok {
		usage()
	}

	dir := beego.AppConfig.String("log_dir")
	clogs.InitLog(dir, "task_"+taskName)

	// 设置进程 title
	procTitle := fmt.Sprintf("%s:%s", programName, taskName)
	gspt.SetProcTitle(procTitle)
	//time.Sleep(100000 * time.Second)

	// 派发任务
	api := taskWork0Map[taskName]

	f := func() {
		str := task.GetCurrentData()
		logs.Warn("[main] handling signal current data:%s", str)
		api.Cancel()
	}

	tools.ClearOnSignal(f)

	api.Start()
}
