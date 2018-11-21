package main

import (
	"fmt"

	// 数据库初始化
	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"

	//"github.com/astaxie/beego/logs"
	"micro-loan/common/models"

	"encoding/json"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

func main() {

	user_e_account := models.User_E_Account{}
	var arr []models.User_E_Account
	var vaCallback models.XenditCallBack

	o := orm.NewOrm()
	o.Using(user_e_account.Using())

	i := 1
	num := 1000

	for {
		offset := (i - 1) * num
		sql := fmt.Sprintf(`select * from user_e_account where va_company_code = 1 limit %d, %d`, offset, num)
		ret, _ := o.Raw(sql).QueryRows(&arr)

		len := len(arr)

		for j := 0; j < len; j++ {
			err := json.Unmarshal([]byte(arr[j].CallbackJson), &vaCallback)
			if err != nil {
				logs.Debug("The id is ", arr[j].Id, ". The err is ", err)
			} else {
				arr[j].BankCode = vaCallback.BankCode
				_, err = arr[j].UpdateEAccount(&arr[j])
				if err != nil {
					logs.Debug("Update faild. ", arr[j].Id)
				} else {
					logs.Debug("Update Successfully. ", arr[j].Id)
				}

			}

		}

		if ret == 0 {
			break
		}
		i++
	}

}
