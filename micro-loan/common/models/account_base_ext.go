package models

import (
	"micro-loan/common/types"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

const ACCOUNT_BASE_EXT_TABLENAME string = "account_base_ext"

type AccountBaseExt struct {
	AccountId                    int64 `orm:"pk"`
	AuthorizeStatusYys           int
	AuthorizeFinishTimeYys       int64
	AuthorizeCrawleTimeYys       int64
	QuotaRaiseYys                int64
	AuthorizeStatusGoJek         int
	AuthorizeFinishTimeGoJek     int64
	AuthorizeCrawleTimeGoJek     int64
	QuotaRaiseGoJek              int64
	AuthorizeStatusLazada        int
	AuthorizeFinishTimeLazada    int64
	AuthorizeCrawleTimeLazada    int64
	QuotaRaiseLazada             int64
	AuthorizeStatusTokopedia     int
	AuthorizeFinishTimeTokopedia int64
	AuthorizeCrawleTimeTokopedia int64
	QuotaRaiseTokopedia          int64
	AuthorizeStatusFacebook      int
	AuthorizeFinishTimeFacebook  int64
	AuthorizeCrawleTimeFacebook  int64
	QuotaRaiseFacebook           int64
	AuthorizeStatusInstagram     int
	AuthorizeFinishTimeInstagram int64
	AuthorizeCrawleTimeInstagram int64
	QuotaRaiseInstagram          int64
	AuthorizeStatusLinkedin      int
	AuthorizeFinishTimeLinkedin  int64
	AuthorizeCrawleTimeLinkedin  int64
	QuotaRaiseLinkedin           int64
	RecallTag                    int
	NpwpNo                       string
	NpwpStatus                   int
	NpwpTime                     int64
	QuotaRaiseNpwp               int64
	PushMessageMark              int64
	PhyInvalidTag                int
	PageAfterLiveFlag            string
	IsManualIdentity             int
	Ctime                        int64
	Utime                        int64
}

func (r *AccountBaseExt) TableName() string {
	return ACCOUNT_BASE_EXT_TABLENAME
}

func (r *AccountBaseExt) Using() string {
	return types.OrmDataBaseApi
}
func (r *AccountBaseExt) UsingSlave() string {
	return types.OrmDataBaseApiSlave
}

func OneAccountBaseExtByPkId(accountId int64) (one AccountBaseExt, err error) {
	o := orm.NewOrm()
	o.Using(one.Using())

	err = o.QueryTable(one.TableName()).
		Filter("account_id", accountId).
		One(&one)
	if err != nil && err != orm.ErrNoRows {
		logs.Error("[OneAccountBaseExtByPkId] err:%v accountId:%d", err, accountId)
	}
	return
}

func (r *AccountBaseExt) InsertWithNoReturn() error {
	// 新增
	_, err := OrmInsert(r)
	if err != nil {
		logs.Error("[InsertWithNoReturn] OrmInsert err:%v, r:%#v", err, r)
	}

	return err
}

func (r *AccountBaseExt) UpdateWithNoReturn(cols []string) error {
	// 新增
	_, err := OrmUpdate(r, cols)
	if err != nil {
		logs.Error("[UpdateWithNoReturn] OrmUpdate err:%v, r:%#v cols:%#v", err, r, cols)
	}
	return err

}

// 查询可召回用户
func GetNeedRecallCustomer(lastAccountID int64) (list []AccountBaseExt, err error) {
	accountBaseExt := AccountBaseExt{}

	o := orm.NewOrm()
	o.Using(accountBaseExt.UsingSlave())

	cond := orm.NewCondition()
	cond = cond.And("recall_tag", types.RecallTagScore)
	cond = cond.And("account_id__gt", lastAccountID)

	_, err = o.QueryTable(accountBaseExt.TableName()).
		SetCond(cond).
		OrderBy("account_id").
		Limit(100).
		All(&list)
	return
}
