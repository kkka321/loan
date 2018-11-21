package main

import (
	"math/rand"
	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"
	"micro-loan/common/pkg/repayremind"
)

func main() {
	// repayremind.FilterAndCreateCases()
	for i := 0; i < 10000; i++ {
		repayremind.PreHandleTest(rand.Int63n(10000000) + 100000000)
	}

}
