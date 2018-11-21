package main

import (
	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"
	"micro-loan/common/pkg/entrust/serveentrust"
)

func main() {
	serveentrust.EntrustRepayList(180823020000009962)
	// result := entrust.GetRepayList("dachuigroup")
	// logs.Notice("result:", result)
}
