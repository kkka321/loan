package main

import (
	"fmt"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"github.com/erikdubbelboer/gspt"

	"micro-loan/common/models"

	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"
	"micro-loan/common/pkg/system/config"
	"micro-loan/common/thirdparty/tongdun"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

func queryData(id int64) (list []models.AccountBase, err error) {
	orderM := models.AccountBase{}
	o := orm.NewOrm()
	o.Using(orderM.Using())

	sql := fmt.Sprintf(`SELECT * FROM %s
WHERE id > %d
LIMIT %d`,
		orderM.TableName(),
		id,
		1000)

	_, err = o.Raw(sql).QueryRows(&list)

	return
}

func main() {
	// 设置进程 title
	procTitle := "fix_gojek"
	gspt.SetProcTitle(procTitle)

	logs.Info("[%s] start launch.", procTitle)

	maxId := int64(0)
	gopoint, _ := config.ValidItemInt("risk_gojek_gopoint")

	for {
		list, _ := queryData(maxId)

		if len(list) == 0 {
			return
		}

		datas := make([]models.AccountBase, 0)

		for _, v := range list {
			if v.Id > maxId {
				maxId = v.Id
			}

			oldV := v.PlatformMark

			_, gojekData, err := tongdun.GetGojekData(v.Id)
			if err != nil {
				continue
			}

			if gojekData.AccountInfo.GojekPoin == "" {
				continue
			}

			point, err := tools.Str2Int(gojekData.AccountInfo.GojekPoin)
			if err != nil {
				continue
			}

			if point > gopoint {
				v.SetPlatformMark(types.PlatformMark_Gojek)
			}

			if oldV == v.PlatformMark {
				continue
			}

			datas = append(datas, v)
		}

		if len(datas) > 0 {
			o := orm.NewOrm()
			m := models.AccountBase{}
			o.Using(m.Using())

			o.Begin()
			for _, v := range datas {
				o.Update(&v, "platform_mark")
			}
			o.Commit()
		}
	}

}
