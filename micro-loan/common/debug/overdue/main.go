package main

import (
	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"
	"micro-loan/common/pkg/system/config"
	"micro-loan/common/service"

	"github.com/astaxie/beego/logs"
)

func main() {
	//overdue.CreateDailySelfUrgeOrders()
	// fmt.Println(overdue.EdgeOrderIsSelfUrge(180508020004666043))
	// fmt.Println(overdue.EdgeMultiOrdersFilterSelfUrge([]int64{180508020004666043, 1, 2, 3, 180511020007376076, 7, 180530020028238360, 9, 10}))

	// cases, _ := entrust.GetEntrustList("180319020025299038,180321020025763211,180723020209149649", 10)
	// cases, _ := entrust.GetEntrustList("", 10)
	// fmt.Println(cases)
	configSimilarVal1, _ := config.ValidItemFloat64("firstenv_reloanenv_similar")
	s, t := service.SaveLoanIDHeadAndLivingEnvCompare(180725010787367663, 180925020000015457)
	configSimilarVal2, _ := config.ValidItemFloat64("first_idhand_reloanenv_similar")

	logs.Debug("sss:", s, "type:", t, "1:", configSimilarVal1, "2:", configSimilarVal2)
}
