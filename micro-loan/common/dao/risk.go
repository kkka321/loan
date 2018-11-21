package dao

import (
	"encoding/json"
	"fmt"

	"github.com/astaxie/beego/orm"

	"micro-loan/common/models"
	"micro-loan/common/types"

	"micro-loan/common/tools"
)

type RiskAdvanceMsg struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Similarity float64 `json:"similarity"`
	} `json:"data"`
}

type RiskFaceidMsg struct {
	Data []struct {
		Quality float64 `json:"quality"`
	} `json:"faces"`
}

func GetFantasyAdvanceResponse(index string, accountId int64) (msg RiskAdvanceMsg) {
	m := models.ThirdpartyRecord{}

	o := orm.NewOrm()
	o.Using(m.UsingSlave())

	sql := fmt.Sprintf(`SELECT * FROM %s
WHERE thirdparty=%d AND related_id=%d AND api LIKE "%%%s"
ORDER BY id DESC LIMIT 0,1`,
		m.TableName(),
		models.ThirdpartyAdvance,
		accountId,
		index,
	)

	list := make([]models.ThirdpartyRecord, 0)
	o.Raw(sql).QueryRows(&list)

	if len(list) == 0 {
		return
	}

	srt := list[0].Response

	err := json.Unmarshal([]byte(srt), &msg)
	if err != nil {
		return
	}

	return
}

func GetFantasyFaceidResponse(index string, accountId int64) (msg RiskFaceidMsg) {
	m := models.ThirdpartyRecord{}

	o := orm.NewOrm()
	o.Using(m.UsingSlave())

	sql := fmt.Sprintf(`SELECT * FROM %s
WHERE thirdparty=%d AND related_id=%d AND api LIKE "%%%s"
ORDER BY id DESC LIMIT 0,1`,
		m.TableName(),
		models.ThirdpartyFaceid,
		accountId,
		index,
	)

	list := make([]models.ThirdpartyRecord, 0)
	o.Raw(sql).QueryRows(&list)

	if len(list) == 0 {
		return
	}

	srt := list[0].Response

	err := json.Unmarshal([]byte(srt), &msg)
	if err != nil {
		return
	}

	return
}

func GetFantasyClientInfo(accountId int64) (list []models.ClientInfo) {
	m := models.ClientInfo{}

	o := orm.NewOrm()
	o.Using(m.UsingSlave())

	infoMap := make(map[string]models.ClientInfo)
	imeiList := make([]string, 0)

	sql := fmt.Sprintf(`SELECT * FROM %s
WHERE related_id=%d AND service_type IN (%d, %d, %d)`,
		m.TableName(),
		accountId,
		types.ServiceRegisterOrLogin, types.ServiceRegister, types.ServiceLogin,
	)

	list1 := make([]models.ClientInfo, 0)
	o.Raw(sql).QueryRows(&list1)
	for _, v := range list1 {
		if _, ok := infoMap[v.Imei]; !ok {
			if v.Imei == "" {
				continue
			}
			imeiList = append(imeiList, "\""+v.ImeiMd5+"\"")

			infoMap[v.Imei] = v
		}
	}

	if len(imeiList) == 0 {
		return
	}

	imeiStr := tools.ArrayToString(imeiList, ",")
	sql = fmt.Sprintf(`SELECT * FROM %s
WHERE related_id!=%d AND imei_md5 IN (%s) AND service_type IN (%d, %d, %d)`,
		m.TableName(),
		accountId,
		imeiStr,
		types.ServiceRegisterOrLogin, types.ServiceRegister, types.ServiceLogin,
	)

	list2 := make([]models.ClientInfo, 0)
	o.Raw(sql).QueryRows(&list2)

	list = append(list, list1...)
	list = append(list, list2...)

	return
}
