package main

import (
	"fmt"

	"micro-loan/common/models"

	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"

	"github.com/astaxie/beego/orm"
	//"micro-loan/common/models"
	//"micro-loan/common/types"
	//"micro-loan/common/thirdparty/advance"
)

func main() {
	var min int

	obj := models.LiveVerify{}

	o := orm.NewOrm()
	o.Using(obj.UsingSlave())

	sqlList := fmt.Sprintf("SELECT *  FROM `%s`  WHERE ctime >= 1527782400000 AND ctime < 1533052800000 ",
		models.LIVE_VERIFY_TABLENAME)
	objs := []models.LiveVerify{}
	num, _ := o.Raw(sqlList).QueryRows(&objs)
	for _, value := range objs {
		avg := (value.ConfidenceRef1 + value.ConfidenceRef2 + value.ConfidenceRef3) / 3
		if avg <= 75 {
			min++
		}
	}
	fmt.Println("------------:", num, min)
}
