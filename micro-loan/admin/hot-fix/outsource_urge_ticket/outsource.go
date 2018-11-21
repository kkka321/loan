package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	// 数据库初始化
	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"
	"micro-loan/common/pkg/ticket"
	"micro-loan/common/types"

	"github.com/astaxie/beego/logs"
	"github.com/erikdubbelboer/gspt"

	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/tools"
)

func main() {

	// 设置进程 title
	procTitle := "set_outsource_for_ticket"
	gspt.SetProcTitle(procTitle)

	logs.Info("[%s] start launch.", procTitle)

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	// +1 分布式锁
	lockKey := fmt.Sprintf("lock:%s", procTitle)
	lock, err := storageClient.Do("SET", lockKey, tools.GetUnixMillis(), "NX")

	if err != nil || lock == nil {
		logs.Error("[%s] process is working, so, I will exit.", procTitle)
		return
	}
	defer storageClient.Do("DEL", lockKey)

	var filePath string

	flag.StringVar(&filePath, "file_path", "", "outsource 数据文件路径,order id以,每行一个")
	flag.Parse()

	if len(filePath) == 0 {
		fmt.Println("file_path is required")
		return
	}
	f, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	succsT := []int64{}
	completeT := []int64{}
	closeT := []int64{}

	rd := bufio.NewReader(f)
	for {
		line, err := rd.ReadBytes('\n') //以'\n'为结束符读入一行
		//line, err := rd.ReadLine() //以'\n'为结束符读入一行
		if err != nil || io.EOF == err {
			break
		}
		if line[len(line)-1] == '\n' {
			drop := 1
			if len(line) > 1 && line[len(line)-2] == '\r' {
				drop = 2
			}
			line = line[:len(line)-drop]
		}
		strLine := string(line)

		orderID, err := strconv.ParseInt(strLine, 10, 64)
		if err != nil {
			fmt.Println(strLine, err)
			continue
		}

		overdueCase, _ := models.OneOverdueCaseByOrderID(orderID)
		if overdueCase.Id == 0 {
			continue
		}
		ticketModel, _ := models.GetTicketByItemAndRelatedID(types.OverdueLevelTicketItemMap()[overdueCase.CaseLevel], overdueCase.Id)
		if ticketModel.Id == 0 {
			continue
		}

		switch ticketModel.Status {
		case types.TicketStatusCompleted:
			// already completed , don't outsource
			completeT = append(completeT, orderID)
			fmt.Println("complete:", orderID, ticketModel)
		case types.TicketStatusClosed:
			// 异常数据
			closeT = append(closeT, orderID)
			fmt.Println("close:", orderID, ticketModel)

		case types.TicketStatusProccessing, types.TicketStatusAssigned, types.TicketStatusCreated:
			ticket.CloseByTicketModel(&ticketModel, "Already Outsourced")
			succsT = append(succsT, orderID)
			fmt.Println("succ:", orderID, ticketModel)
		}
		time.Sleep(time.Millisecond * 10)
	}

	fmt.Println("succsT:", succsT)
	fmt.Println("completeT:", completeT)
	fmt.Println("closeT:", closeT)

}
