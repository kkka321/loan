package main

import (
	//"encoding/json"

	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"
	"micro-loan/common/service"
	//"micro-loan/common/thirdparty/advance"
	//"micro-loan/common/thirdparty/faceid"
	//"micro-loan/common/thirdparty/textlocal"
	"github.com/astaxie/beego/logs"
)

func main() {

	arr := []int{1, 2, 3, 4, 5, 6, 7}
	logs.Debug(arr[2:4])
	logs.Debug(arr[2:4])
	logs.Debug(arr[4:6])

	isdone, edu, baifenbi, authList := service.CustomerAuthorize(180301010007362546)
	logs.Notice("isdone:", isdone, "edu:", edu, "baifenbi:", baifenbi, "authList:", authList)

}
