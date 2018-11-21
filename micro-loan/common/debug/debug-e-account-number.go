package main

import (
	//"encoding/json"

	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"

	"micro-loan/common/service"
	//"micro-loan/common/thirdparty/advance"
	//"micro-loan/common/thirdparty/faceid"
	//"micro-loan/common/thirdparty/textlocal"
)

func main() {

	data := map[string]interface{}{
		"test": "test",
	}

	service.ApiDataAddEAccountNumber(180508010004732233, data)
}
