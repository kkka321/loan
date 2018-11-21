package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"micro-loan/common/lib/sms"
	"micro-loan/common/lib/sms/api"
	"micro-loan/common/models"
	"micro-loan/common/tools"
	"micro-loan/common/types"
	"os"
	"strings"

	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"

	"github.com/astaxie/beego/logs"
	"github.com/erikdubbelboer/gspt"
)

var msgFileName string
var listFileName string

func init() {
	flag.StringVar(&msgFileName, "msg", "", "msg name")
	flag.StringVar(&listFileName, "list", "", "list name")
}

func main() {
	flag.Parse()

	// 设置进程 title
	procTitle := "send_sms"
	gspt.SetProcTitle(procTitle)

	logs.Info("[%s] start launch.", procTitle)

	buf, err := ioutil.ReadFile(msgFileName)
	if err != nil {
		fmt.Println("ReadFile msg:", err)
		return
	}

	msgStr := string(buf)
	fmt.Println("msg:", msgStr)

	f, err := os.Open(listFileName)
	if err != nil {
		fmt.Println("ReadFile listFileName:", err)
		return
	}
	defer f.Close()

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

		accountIdStr := strings.Trim(vec[0], " ")
		if accountIdStr == "" {
			continue
		}

		accountId, _ := tools.Str2Int64(accountIdStr)
		fmt.Println("orderId:", accountId)

		accountBase, err := models.OneAccountBaseByPkId(accountId)
		if err != nil {
			fmt.Println("OneAccountBaseByPkId err:", err)
			continue
		}

		sms.SendByKey(api.Sms253ID, types.ServiceMonitor, accountBase.Mobile, msgStr, accountId)
	}
}
