package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/astaxie/beego/logs"
	"github.com/erikdubbelboer/gspt"

	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"

	"micro-loan/common/pkg/task"
	"micro-loan/common/tools"
	//"micro-loan/common/types"
)

var fileName string

func init() {
	flag.StringVar(&fileName, "name", "", "file name")
}

func main() {
	flag.Parse()

	// 设置进程 title
	procTitle := "offline-risk"
	gspt.SetProcTitle(procTitle)

	logs.Info("[%s] start launch.", procTitle)

	f, err := os.Open(fileName)
	if err != nil {
		logs.Error("[%s] no input file, fileName: %s, err: %s", procTitle, fileName, err.Error())
		fmt.Printf("Usage: ./%s --name=input\n", procTitle)
		return
	}
	defer f.Close()

	t := time.Now()
	outputFile := fmt.Sprintf("./offline-risk-result.%d%02d%02d-%04d",
		t.Year(), t.Month(), t.Day(), t.Unix()%10000)
	output, err := os.Create(outputFile)
	if err != nil {
		logs.Error("[%s] can open file: %s, err: %s", procTitle, outputFile, err.Error())
		os.Exit(20)
	}
	defer output.Close()

	rd := bufio.NewReader(f)
	for {
		line, err := rd.ReadString('\n') //以'\n'为结束符读入一行
		if err != nil {
			if io.EOF == err {
				logs.Informational("[%s] file get EOF", procTitle)
			} else {
				logs.Error("[%s] read file has wrong, err: %s", procTitle, err.Error())
			}

			break
		}

		line = strings.TrimRight(line, "\n")
		vec := strings.Split(line, "\t")
		if len(vec) < 1 {
			logs.Warning("[%s] split data has wrong, line: %s", procTitle, line)
			continue
		}

		accountID, err := tools.Str2Int64(vec[0])
		if err != nil {
			logs.Warning("[%s] unexpected data: %s", procTitle, line)
			continue
		}

		hitRisk := task.OfflineHandleRiskReview(accountID)
		var hits []string
		for _, riskItem := range hitRisk {
			hits = append(hits, riskItem.Regular)
		}
		outLine := fmt.Sprintf("%d\t%s\n", accountID, strings.Join(hits, ", "))
		output.WriteString(outLine)
	}

	logs.Info("[%s] jobs have done.", procTitle)
}
