package main

import (
	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"
	"micro-loan/common/pkg/event"
	"micro-loan/common/pkg/event/evtypes"
)

func main() {
	// repayremind.FilterAndCreateCases()
	event.Trigger(&evtypes.FixPaymentCodeEv{OrderID: 1111111111})

}
