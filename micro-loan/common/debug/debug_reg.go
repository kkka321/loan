package main

import (
	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"
	"micro-loan/common/tools"

	"micro-loan/common/thirdparty/doku"

	"github.com/astaxie/beego/logs"
)

func main() {
	valid := tools.ContainNumber("asdfasfas1asdf")
	logs.Debug(valid)

	bankCode, _ := doku.GetDoKuDisburseVABankCode("Bank CIMB Niaga")
	logs.Debug(bankCode)
	prefix := doku.GetDokuVAPrefix(bankCode)
	logs.Debug(prefix)
}
