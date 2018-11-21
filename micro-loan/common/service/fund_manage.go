package service

import (
	"fmt"
	"strings"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"

	"micro-loan/common/models"
	"micro-loan/common/types"
)

var bankFiledNameMap = map[int]string{
	types.Xendit: types.XenditFiledName,
	//types.Bluepay: types.BluepayFiledName,
	types.DoKu: types.DokuFiledName,
}

func getCompanyTypeFiledName(loanRepayType int) string {
	switch loanRepayType {
	case types.LoanRepayTypeLoan:
		{
			return "loan_company_code"
		}
	case types.LoanRepayTypeRepay:
		{
			return "repay_company_code"
		}
	default:
		{
			logs.Warn("[getCompanyTypeFiledName] loanRepayType:%d", loanRepayType)
			return ""
		}
	}
}

func BankList(compaynCode int, loanRepayType int) (assignedList []models.BanksInfo, unAssignedList []models.BanksInfo, allUnAssignedList []models.BanksInfo, err error) {

	briveFiledName := bankFiledNameMap[compaynCode]
	if len(briveFiledName) == 0 {
		err = fmt.Errorf("[BankList] unknow companyCode :%d", compaynCode)
		return
	}

	obj := models.BanksInfo{}
	o := orm.NewOrm()
	o.Using(obj.UsingSlave())

	// 初始化查询条件
	// assign condition
	condAssign := orm.NewCondition()
	if loanRepayType != types.LoanRepayTypeRepay {
		condAssign = condAssign.AndNot(briveFiledName, "")
	}

	// unassign condition
	condUnAssign := orm.NewCondition()
	if loanRepayType != types.LoanRepayTypeRepay {
		condUnAssign = condUnAssign.AndNot(briveFiledName, "")
	}

	// all unassign condition
	condAllUnAssign := orm.NewCondition()

	companyTypeName := getCompanyTypeFiledName(loanRepayType)
	condAssign = condAssign.And(companyTypeName, compaynCode)
	condUnAssign = condUnAssign.And(companyTypeName, types.None)
	condAllUnAssign = condAllUnAssign.And(companyTypeName, types.None)

	_, err = o.QueryTable(obj.TableName()).SetCond(condAssign).OrderBy("full_name").All(&assignedList)
	if err != nil {
		logs.Error("[BankList] query assigned err:%v loanRepayType:%d", err, loanRepayType)
		return
	}

	_, err = o.QueryTable(obj.TableName()).SetCond(condUnAssign).OrderBy("full_name").All(&unAssignedList)
	if err != nil {
		logs.Error("[BankList] query condUnAssign err:%v loanRepayType:%d", err, loanRepayType)
		return
	}

	_, err = o.QueryTable(obj.TableName()).SetCond(condAllUnAssign).OrderBy("full_name").All(&allUnAssignedList)
	if err != nil {
		logs.Error("[BankList] query allUnAssignedList err:%v loanRepayType:%d", err, loanRepayType)
		return
	}
	return
}

func BankAssign(compaynCode int, loanRepayType int, assignOperations []string) (err error) {

	obj := models.BanksInfo{}
	o := orm.NewOrm()
	o.Using(obj.Using())

	filedName := getCompanyTypeFiledName(loanRepayType)
	assignIds := strings.Join(assignOperations, ",")
	sql := "update %s set %s = %d where id in(%s)"
	sql = fmt.Sprintf(sql, obj.TableName(), filedName, compaynCode, assignIds)
	logs.Info("[BankAssign] sql:%s", sql)

	num := 0
	o.Raw(sql).QueryRow(&num)
	return
}

func BankUnAssign(compaynCode int, loanRepayType int, assignOperations []string) (err error) {
	if len(assignOperations) == 0 {
		logs.Error("[BankUnAssign] len(assignOperations) == 0 . loanRepayType:%d", loanRepayType)
		return
	}

	obj := models.BanksInfo{}
	o := orm.NewOrm()
	o.Using(obj.Using())

	filedName := getCompanyTypeFiledName(loanRepayType)
	assignIds := strings.Join(assignOperations, ",")
	sql := "update %s set %s = %d where id in(%s) and %s = %d"
	sql = fmt.Sprintf(sql, obj.TableName(), filedName, types.None, assignIds, filedName, compaynCode)
	logs.Info("[BankUnAssign] sql:%s", sql)

	num := 0
	o.Raw(sql).QueryRow(&num)
	return
}
