package main

import (
	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"

	"github.com/astaxie/beego/logs"
)

func main() {

	// ch := make(chan string, 3)
	// ch <- "A"
	// ch <- "B"
	// ch <- "C"

	// logs.Notice("[ch]:", <-ch)
	// logs.Notice("[ch]:", <-ch)
	// logs.Notice("[ch cap]:", cap(ch))
	// logs.Notice("[ch len]:", len(ch))

	type Apple struct {
		Color string
		Size  int
	}
	ch := make(chan Apple, 3)

	go func() {
		ch <- Apple{Color: "red", Size: 2}
	}()
	go func() {
		ch <- Apple{Color: "green", Size: 2}
	}()
	go func() {
		ch <- Apple{Color: "blue", Size: 2}
	}()

	// ch <- Apple{Color: "green", Size: 3}
	// ch <- Apple{Color: "blue", Size: 4}

	logs.Notice("[ch]:", <-ch)
	logs.Notice("[ch]:", <-ch)
	logs.Notice("[ch cap]:", cap(ch))
	logs.Notice("[ch len]:", len(ch))
}
