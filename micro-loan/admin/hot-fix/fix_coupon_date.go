package main

import (
	"fmt"

	"time"

	_ "micro-loan/common/lib/db/mysql"

	"github.com/astaxie/beego/orm"

	"micro-loan/common/models"
)

func main() {
	couponSql := "select * from coupon"

	m := models.Coupon{}
	o := orm.NewOrm()
	o.Using(m.Using())
	couponList := make([]models.Coupon, 0)
	o.Raw(couponSql).QueryRows(&couponList)

	count := int64(0)
	for _, v := range couponList {
		for {
			sql := fmt.Sprintf(`UPDATE account_coupon SET valid_start=%d, valid_end=%d
WHERE coupon_id=%d AND (valid_start=0 OR valid_end=0)
LIMIT 500`,
				v.ValidStart, v.ValidEnd,
				v.Id)

			r, e := o.Raw(sql).Exec()
			if e != nil {
				fmt.Sprintln("count:%d, Raw err:%v", count, e)
				break
			}

			c, e := r.RowsAffected()
			if e != nil {
				fmt.Sprintln("count:%d, RowsAffected err:%v", count, e)
				break
			}

			count += c

			if c < 500 {
				fmt.Sprintln("count:%d, RowsAffected rows:%d", c)
				break
			}

			fmt.Sprintln("count:%d", count)
			time.Sleep(time.Millisecond * 10)
		}
	}
}
