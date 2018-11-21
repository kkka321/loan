package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/astaxie/beego/logs"
	"github.com/erikdubbelboer/gspt"

	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"

	"micro-loan/common/pkg/google/push"
	"micro-loan/common/tools"
	"micro-loan/common/types"
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
	procTitle := "push_message"
	gspt.SetProcTitle(procTitle)

	logs.Info("[%s] start launch.", procTitle)

	msgStr, err := ioutil.ReadFile(msgFileName)
	if err != nil {
		fmt.Println("ReadFile msg:", err)
		return
	}

	msgL := strings.Split(string(msgStr), "\n")
	if len(msgL) < 2 {
		fmt.Println("Split msg:", string(msgStr))
		return
	}

	title := msgL[0]
	msg := msgL[1]

	fmt.Println("title:", title)
	fmt.Println("msg:", msg)

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

		push.SendFmsMessage(accountId, title, msg, types.MessageTypeOther)
	}
}
