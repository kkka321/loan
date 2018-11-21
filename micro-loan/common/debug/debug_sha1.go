package main

import (
	"crypto/sha1"
	"encoding/hex"

	"micro-loan/common/tools"

	"github.com/astaxie/beego/logs"
)

func Sha1(data string) string {
	sha1 := sha1.New()
	sha1.Write([]byte(data))
	return hex.EncodeToString(sha1.Sum([]byte("")))
}

func main() {
	// var orderID, userAccountID int64 = 180523020015578858, 180404010035151464
	// ticket.CreateTicket(types.TicketItemRepayRemind, 22222, types.Robot,
	// 	ticket.DataRepayRemindCase{OrderID: orderID, CustomerID: userAccountID})

	// ticket.CreateTicket(types.TicketItemPhoneVerify, 180523020015578858, types.Robot,
	// 	ticket.DataRepayRemindCase{OrderID: orderID, CustomerID: userAccountID})
	//event.Trigger(&evtypes.TicketCreateEv{types.TicketItemPhoneVerify, types.Robot, 180511020007458399, nil})

	//ticket.WorkerIncompletedTicketsByTicketItem(types.TicketItemPhoneVerify)
	//result, err := ticket.IdleAssign(types.TicketItemPhoneVerify)
	// logs.Debug("Assign user:", result)
	// logs.Debug(err)

	logs.Debug(Sha1("10000.005870TFwpA430hJ4n1396430482839SUCCESSNA"))

	va := "1234567890123456"

	logs.Debug(va[0:4])
	logs.Debug(va[4:8])
	logs.Debug(va[8:12])
	logs.Debug(va[12:])

	s1 := tools.SubString(va, 0, 4)
	logs.Debug(s1)
	s2 := tools.SubString(va, 4, 4)
	logs.Debug(s2)
	s3 := tools.SubString(va, 8, 4)
	logs.Debug(s3)
	s4 := tools.SubString(va, 12, 4)
	logs.Debug(s4)
}
