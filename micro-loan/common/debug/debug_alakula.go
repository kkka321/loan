package main

import (
	_ "micro-loan/common/models"
	_ "micro-loan/common/types"
	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"
	_ "micro-loan/common/service"
	_ "github.com/astaxie/beego/logs"
	_ "github.com/astaxie/beego"
	_ "encoding/json"
	"micro-loan/common/thirdparty/Akulaku"
)

func main() {
	test2("Doe2hU0eThvrcDfr", "8NkHjPUJ0LzAfRC6", "TUHAN", "3510243006730604")
}


func test2(secretKey string, appkey string, name string, ktp string) {
	Akulaku.CheckRisk("WAWANSETIAWAN", "3210101909870021")

	Akulaku.CheckRisk("TUHAN", "3510243006730604")

	Akulaku.CheckRisk("MUHAMI", "3624041103830012")
}
