package main

import (
	// 数据库初始化
	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"github.com/erikdubbelboer/gspt"

	"bufio"
	"flag"
	"fmt"
	"io"
	"micro-loan/common/models"
	"micro-loan/common/service"
	"micro-loan/common/tools"
	"micro-loan/common/types"
	"os"
	"strings"
)

var fileName string

func init() {
	flag.StringVar(&fileName, "name", "", "file name")
}

func main() {
	flag.Parse()

	// 设置进程 title
	procTitle := "upload_blacklist"
	gspt.SetProcTitle(procTitle)

	logs.Info("[%s] start launch.", procTitle)

	m := models.CustomerRisk{}
	o := orm.NewOrm()
	o.Using(m.Using())

	var list []models.CustomerRisk
	o.QueryTable(m.TableName()).All(&list)

	var newDatas map[types.RiskItemEnum]map[string]bool = make(map[types.RiskItemEnum]map[string]bool)
	var oldDatas map[types.RiskItemEnum]map[string]bool = make(map[types.RiskItemEnum]map[string]bool)
	for _, v := range list {
		if v.IsDeleted == 1 {
			continue
		}

		if v.RiskType != 1 {
			continue
		}

		if oldDatas[v.RiskItem] == nil {
			oldDatas[v.RiskItem] = make(map[string]bool)
		}
		oldDatas[v.RiskItem][v.RiskValue] = true
	}

	f, err := os.Open(fileName)
	if err != nil {
		return
	}
	defer f.Close()

	rd := bufio.NewReader(f)
	isFirst := true
	for {
		line, err := rd.ReadString('\n') //以'\n'为结束符读入一行
		if err != nil || io.EOF == err {
			break
		}

		if isFirst {
			isFirst = false
			continue
		}

		vec := strings.Split(line, ",")
		if len(vec) < 2 {
			continue
		}

		accountStr := strings.Trim(vec[0], " ")
		orderIdStr := strings.Trim(vec[1], " ")
		if accountStr == "" || orderIdStr == "" {
			continue
		}

		accountId, _ := tools.Str2Int64(accountStr)
		orderId, _ := tools.Str2Int64(orderIdStr)

		o, err := models.GetOrder(orderId)
		if err != nil {
			continue
		}

		if o.UserAccountId != accountId {
			continue
		}

		accountBase, _ := models.OneAccountBaseByPkId(accountId)
		aclientInfo, err := service.OrderClientInfo(orderId)

		if accountBase.Mobile != "" {
			if _, ok := oldDatas[types.RiskItemMobile][accountBase.Mobile]; !ok {
				if newDatas[types.RiskItemMobile] == nil {
					newDatas[types.RiskItemMobile] = make(map[string]bool)
				}
				newDatas[types.RiskItemMobile][accountBase.Mobile] = true
			}
		}

		if accountBase.Identity != "" {
			if _, ok := oldDatas[types.RiskItemIdentity][accountBase.Identity]; !ok {
				if newDatas[types.RiskItemIdentity] == nil {
					newDatas[types.RiskItemIdentity] = make(map[string]bool)
				}
				newDatas[types.RiskItemIdentity][accountBase.Identity] = true
			}
		}

		if aclientInfo.Imei != "" {
			if _, ok := oldDatas[types.RiskItemIMEI][aclientInfo.Imei]; !ok {
				if newDatas[types.RiskItemIMEI] == nil {
					newDatas[types.RiskItemIMEI] = make(map[string]bool)
				}
				newDatas[types.RiskItemIMEI][aclientInfo.Imei] = true
			}
		}
	}

	timeNow := tools.GetUnixMillis()

	o.Using(m.Using())
	o.Begin()

	var err2 error
	for k, v := range newDatas {
		for k1, _ := range v {
			c := models.CustomerRisk{}
			c.CustomerId = 0
			c.RiskItem = k
			c.RiskType = 1
			c.RiskValue = k1
			c.Status = 1
			c.Reason = 5
			c.Ctime = timeNow
			c.Utime = timeNow
			c.OpUid = 1
			c.ReviewTime = timeNow

			_, err2 = o.Insert(&c)
			if err2 != nil {
				fmt.Println(err2)
				break
			}
		}
	}

	if err2 != nil {
		o.Rollback()
	} else {
		o.Commit()
	}

	logs.Info("[%s] politeness exit.", procTitle)
}
