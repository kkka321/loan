package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"github.com/erikdubbelboer/gspt"

	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"

	"micro-loan/common/models"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

var fileName string

func init() {
	flag.StringVar(&fileName, "name", "", "file name")
}

func main() {
	flag.Parse()

	// 设置进程 title
	procTitle := "upload_blacklist"
	gspt.SetProcTitle(procTitle)

	logs.Info("[%s] start launch.", procTitle)

	m := models.Order{}
	o := orm.NewOrm()
	o.Using(m.Using())

	f, err := os.Open(fileName)
	if err != nil {
		return
	}
	defer f.Close()

	timetag := tools.GetUnixMillis()

	rd := bufio.NewReader(f)
	for {
		line, err := rd.ReadString('\n') //以'\n'为结束符读入一行
		if err != nil || io.EOF == err {
			break
		}

		line = strings.TrimRight(line, "\n")

		vec := strings.Split(line, ",")
		if len(vec) < 1 {
			continue
		}

		orderIdStr := strings.Trim(vec[0], " ")
		if orderIdStr == "" {
			continue
		}

		orderId, _ := tools.Str2Int64(orderIdStr)
		fmt.Println("orderId:%d", orderId)

		order, err := models.GetOrder(orderId)
		if err != nil {
			fmt.Println("orderId:%d not found", orderId)
			continue
		}

		if order.CheckStatus != types.LoanStatusReject {
			fmt.Println("orderId:%d status:%d skip", orderId, order.CheckStatus)
			continue
		}

		order.RiskCtlRegular = ""
		order.CheckStatus = types.LoanStatus4Review
		order.Utime = timetag
		models.UpdateOrder(&order)
	}
}
