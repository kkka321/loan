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
	"micro-loan/common/tools"
	"micro-loan/common/types"
	"os"
	"strings"
)

var fileName string

func init() {
	flag.StringVar(&fileName, "name", "", "file name")
}

type AcctInfo struct {
	mark      string
	accountId int64
}

func main() {
	flag.Parse()

	// 设置进程 title
	procTitle := "upload_account_blacklist"
	gspt.SetProcTitle(procTitle)

	logs.Info("[%s] start launch.", procTitle)

	m := models.CustomerRisk{}
	o := orm.NewOrm()
	o.Using(m.Using())

	var list []models.CustomerRisk
	o.QueryTable(m.TableName()).All(&list)

	var newDatas map[types.RiskItemEnum]map[string]AcctInfo = make(map[types.RiskItemEnum]map[string]AcctInfo)
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

	var clientInfo models.ClientInfo
	o.Using(clientInfo.Using())

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
		if len(vec) < 3 {
			continue
		}

		accountStr := strings.Trim(vec[0], " ")
		accountId, _ := tools.Str2Int64(accountStr)

		accountBase, _ := models.OneAccountBaseByPkId(accountId)
		remark := strings.Trim(vec[2], " ")
		remark = strings.Trim(remark, "\r\n")

		var clients []models.ClientInfo
		o.QueryTable(clientInfo.TableName()).Filter("related_id", accountId).Filter("service_type", types.ServiceRegisterOrLogin).All(&clients)

		if accountBase.Mobile != "" {
			if _, ok := oldDatas[types.RiskItemMobile][accountBase.Mobile]; !ok {
				if newDatas[types.RiskItemMobile] == nil {
					newDatas[types.RiskItemMobile] = make(map[string]AcctInfo)
				}
				var acct AcctInfo = AcctInfo{remark, accountId}
				newDatas[types.RiskItemMobile][accountBase.Mobile] = acct
			}
		}

		if accountBase.Identity != "" {
			if _, ok := oldDatas[types.RiskItemIdentity][accountBase.Identity]; !ok {
				if newDatas[types.RiskItemIdentity] == nil {
					newDatas[types.RiskItemIdentity] = make(map[string]AcctInfo)
				}
				var acct AcctInfo = AcctInfo{remark, accountId}
				newDatas[types.RiskItemIdentity][accountBase.Identity] = acct
			}
		}

		for _, v := range clients {
			if v.Imei != "" {
				if _, ok := oldDatas[types.RiskItemIMEI][v.Imei]; !ok {
					if newDatas[types.RiskItemIMEI] == nil {
						newDatas[types.RiskItemIMEI] = make(map[string]AcctInfo)
					}
					var acct AcctInfo = AcctInfo{remark, accountId}
					newDatas[types.RiskItemIMEI][v.Imei] = acct
				}
			}
		}
	}

	timeNow := tools.GetUnixMillis()

	logs.Info("[%s] begin insert datas", procTitle)

	o.Using(m.Using())
	o.Begin()

	count := 0
	var err2 error
	for k, v := range newDatas {
		for k1, v1 := range v {
			c := models.CustomerRisk{}
			c.CustomerId = v1.accountId
			c.RiskItem = k
			c.RiskType = 1
			c.RiskValue = k1
			c.Status = 1
			c.Reason = types.RiskReasonLiar
			c.Ctime = timeNow
			c.Utime = timeNow
			c.OpUid = 0
			c.ReviewTime = timeNow
			c.ReportRemark = v1.mark
			count++

			_, err2 = o.Insert(&c)
			if err2 != nil {
				fmt.Println(err2)
				break
			}
		}
	}

	logs.Info("[%s] insert done size:%d", procTitle, count)

	if err2 != nil {
		o.Rollback()
		logs.Error(err2)
	} else {
		o.Commit()
	}

	logs.Info("[%s] politeness exit.", procTitle)
}
