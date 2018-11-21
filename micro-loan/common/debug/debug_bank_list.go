package main

import (
	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"
	"micro-loan/common/thirdparty/xendit"

	"github.com/astaxie/beego/logs"
)

func main() {
	bankList := xendit.AllBankList()
	logs.Debug(bankList)
}
