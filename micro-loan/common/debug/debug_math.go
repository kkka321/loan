package main

import (
	"micro-loan/common/pkg/repayplan"

	"github.com/astaxie/beego/logs"
)

func main() {

	val := repayplan.CaculateCanReducedAmount(100, 0.5)
	logs.Debug(val)
}
