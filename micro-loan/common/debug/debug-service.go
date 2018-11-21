package main

import (
	"fmt"

	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"
	"micro-loan/common/models"

	"micro-loan/common/service"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

func main() {
	//testGetValidSystemConfigItemValue()
	////testRoleUser()
	//testSameCompanyApplyLoanOrderInLastMonth("12345678901234567890")
	//testSameResidentAddressApplyLoanOrderInLastMonth("12345678901234567890")
	//testGetTodayLoanOrderTotal()
	//overdueDays()
	// testContactsMaxOverdueDaysInLoanHistory()
	//testContactsOverdueLoanOrderStat()
	// testSameContactsCustomerOverdueStat()
	// testFindCommonContact()
	//testSameResidenceOverdueStat()
	//testSameCompanyOverdueStat()
	//testCustomerOverdueTotalStat()
	//testGetUnservicedAreaConf()
	// testUpdateAccountOtherInfo()
	//testMenuControByOrderStatus()
	// testUpdateAccountBaseOCR()
	//testUpdateAccountBaseByThird()
	testGetEaccount()
}

func testUpdateAccountBaseByThird() {
	// 180211010000040290
	accountBase, _ := models.OneAccountBaseByPkId(180211010000040290)
	accountBase.ThirdID = "22222"
	accountBase.ThirdName = "fff"
	service.UpdateAccountBaseByThird(accountBase)
}

func testUpdateAccountBaseOCR() {

	service.UpdateAccountBaseOCR(180209010000014072, "", "")
	// service.UpdateAccountProfileIdPhoto(180209010000014072, 0, 0)
	testUpdateAccountOtherInfo()
}

func testMenuControByOrderStatus() {

	show := service.MenuControlByOrderStatus(180227010003866725)
	fmt.Print(show)
}

func testUpdateAccountOtherInfo() {
	service.UpdateAccountOtherInfo(180227010003866725, 1, 2, 3, "4", "5", "6", "7")
}

// func testGetValidSystemConfigItemValue() {
// 	itemName := "risk_ctl_D006"
// 	itemValue, err := service.SystemConfigValidItemInt(itemName)
// 	fmt.Printf("[SystemConfigValidItemInt] itemName: %d, err: %v\n", itemValue, err)
// }

func testGetTodayLoanOrderTotal() {
	total, err := service.GetTodayLoanOrderTotal()
	fmt.Printf("total: %d, err: %v\n", total, err)
}

func testSameCompanyApplyLoanOrderInLastMonth(company string) {
	num, err := service.SameCompanyApplyLoanOrderInLastMonth(company)
	fmt.Printf("[SameCompanyApplyLoanOrderInLastMonth] num: %d, err: %v\n", num, err)
}

func testSameResidentAddressApplyLoanOrderInLastMonth(address string) {
	num, err := service.SameResidentAddressApplyLoanOrderInLastMonth(address)
	fmt.Printf("[SameResidentAddressApplyLoanOrderInLastMonth] num: %d, err: %v\n", num, err)
}

func overdueDays() {
	var repayDate int64 = tools.NaturalDay(3)
	level, days, err := service.CalculateOverdueLevel(repayDate)
	fmt.Printf("level: %s, days: %d, err: %#v\n", level, days, err)

	level = types.OverdueLevelM11
	preLevel, err := types.GetPreviousOverdueLevel(level)
	fmt.Printf("level: %s, preLevel: %s, err: %#v\n", level, preLevel, err.Error())
}

func testRoleUser() {
	service.GetUsersByRoleName("TEST")
}

func testContactsMaxOverdueDaysInLoanHistory() {
	mobile := "8613600000096"
	days, err := service.ContactsMaxOverdueDaysInLoanHistory(mobile)
	fmt.Printf("days: %d, err: %#v\n", days, err)
}

func testContactsOverdueLoanOrderStat() {
	mobile := "8613600000096"
	total, err := service.ContactsOverdueLoanOrderStat(mobile)
	fmt.Printf("total: %d, err: %#v\n", total, err)

}

func testSameContactsCustomerOverdueStat() {

	ct1 := "0811000001"
	ct2 := "0811000002"
	var exclude int64 = 180319010025249492
	accountIDs, total, err := service.SameContactsCustomerOverdueStat(ct1, ct2, exclude)
	fmt.Printf("accountIDs:\n")
	fmt.Print(accountIDs)
	fmt.Println("end>>")
	fmt.Printf("[SameContactsCustomerOverdueStat] total: %d, err: %#v\n", total, err)
}

func testFindCommonContact() {

	accountIDs := []int64{180305010009672585, 180305010010048134, 180305010010169831}
	conmmonContact := service.FindCommonContact(accountIDs)
	fmt.Println("commonContact:")
	fmt.Println(conmmonContact)
}

func testSameResidenceOverdueStat() {
	residentCity := "Beijing Shi,Beijing"
	residentAddress := "daxing"
	accountIDs, total, err := service.SameResidenceOverdueStat(residentCity, residentAddress)
	fmt.Printf("accountIDs:\n", accountIDs)
	fmt.Printf("[SameResidenceOverdueStat] total: %d, err: %#v\n", total, err)
}

func testSameCompanyOverdueStat() {
	companyName := "12345678901234567890"
	accountIDs, total, err := service.SameCompanyOverdueStat(companyName)
	fmt.Printf("accountIDs:\n", accountIDs)
	fmt.Printf("[SameCompanyOverdueStat] total: %d, err: %#v\n", total, err)
}

func testCustomerOverdueTotalStat() {
	condBox := map[string]interface{}{
		//"is_overdue": true,
		"check_status":  true,
		"last_3_months": true,
	}
	total, err := service.CustomerOverdueTotalStat(condBox)
	fmt.Printf("[CustomerOverdueTotalStat] total: %d, err: %#v\n", total, err)
}

func testGetUnservicedAreaConf() {
	conf, err := service.GetUnservicedAreaConf()
	fmt.Printf("[GetUnservicedAreaConf] conf: %#v, err: %#v\n", conf, err)
}
