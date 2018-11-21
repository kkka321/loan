package main

import (
	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"
	"micro-loan/common/service"
	"micro-loan/common/tools"
)

func main() {
	//e := service.LoanSubmitEv{1111, 22222}
	//event.Trigger(&evtypes.LoanSubmitEv{OrderID: 180411020037452557, Time: tools.GetUnixMillis()})
	//service.GetOneEvent()
	//service.DoAddBlacklist(180227010003866725, 3, 1, "1111166661")

	timetag := tools.NaturalDay(0)
	service.CustomerRecallTag(timetag)
}
