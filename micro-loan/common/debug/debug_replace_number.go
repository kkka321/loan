package main

import (
	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"

	"micro-loan/common/tools"

	"github.com/astaxie/beego/logs"
)

func main() {

	example := "#Go22Lang   1Codeasdfasd!$!  ___+++@%^&  "
	processedString := tools.TrimRealName(example)
	logs.Debug(example)
	logs.Debug(processedString)

}
