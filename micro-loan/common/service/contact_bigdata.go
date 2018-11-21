package service

import (
	"fmt"
	"micro-loan/common/models"
	"micro-loan/common/tools"
	"sort"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

type ContactList struct {
	Mobile     string // 通讯录手机号
	Name       string // 通讯录姓名
	Num        int    // 拨打次数
	IsDailLast int    // 通话记录中上次是否拨通, -1:未拨打过, 0:未拨通, 1:拨通
}

type ContactListType int

const (
	ContactListUrge  ContactListType = 0
	ContactListRepay ContactListType = 1
)

// 电话拨打状态
const (
	DailStatusNotCall   int = -1 // 未拨打过
	DailStatusNotCalled int = 0  // 未拨通
	DailStatusCalled    int = 1  // 拨通
)

func GetContactList(t ContactListType, accountId, orderId int64) (hasBigData bool, contactList []ContactList) {

	objs, num, err := models.OneAccountBigdataContactByAccountID(accountId)
	if err != nil {
		logs.Error("[GetContactList] Get contact list error:", err, ", accountId:", accountId)
		return
	}
	if num <= 0 {
		logs.Info("[GetContactList] Get contact list is blank, accountId:", accountId)
		return
	}

	hasBigData = true
	callRecord := true
	var bigData EsResponse
	clientInfo, err := models.OneLastClientInfoByRelatedID(accountId)
	if err != nil {
		callRecord = false
	} else {
		bigData, _, _, err = EsSearchById(tools.Md5(clientInfo.Imei))
		if err != nil || bigData.Found != true {
			callRecord = false
		}
	}

	for _, v := range objs {
		var contact ContactList
		contact.Mobile = tools.MobileFormat(v.Mobile)
		contact.Name = v.ContactName
		contact.IsDailLast = GetPhoneConnectFromRecord(t, orderId, v.Mobile)
		contact.Num = 0
		if callRecord {
			d := bigData.Source.NumberOfCallsToFirstContact
			if n, ok := d[contact.Mobile]; ok {
				contact.Num = n
			}
		}
		contactList = append(contactList, contact)
	}

	contactList = ContactListSortByNum(contactList)

	return
}

// []ContactList 按照 num 排序
func ContactListSortByNum(c []ContactList) []ContactList {

	sort.Slice(c, func(i, j int) bool {
		return c[i].Num > c[j].Num
	})

	return c
}

func GetPhoneConnectFromRecord(t ContactListType, orderId int64, mobile string) (isDail int) {
	//var contact ContactList
	var tableName string
	o := orm.NewOrm()

	if t == ContactListUrge {
		obj := models.OverdueCaseDetail{}
		o.Using(obj.Using())
		tableName = obj.TableName()
	} else {
		obj := models.RepayRemindCaseLog{}
		o.Using(obj.Using())
		tableName = obj.TableName()
	}

	// 初始化查询条件
	selectSql := fmt.Sprintf(`SELECT phone_connect`)
	where := fmt.Sprintf(`where order_id = %d and phone_object_mobile = %s`, orderId, mobile)
	sqlList := fmt.Sprintf(`%s FROM %s %s ORDER BY id desc limit 1`, selectSql, tableName, where)

	// 查询指定页
	r := o.Raw(sqlList)
	err := r.QueryRow(&isDail)
	if err != nil {
		// logs.Info("[GetPhoneConnectFromRecord] not exist record, mobile:%s", mobile)
		isDail = DailStatusNotCall
	}

	return
}
