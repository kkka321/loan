package models

import (
	"fmt"
	"time"

	//"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"

	"micro-loan/common/lib/device"
	"micro-loan/common/tools"
	"micro-loan/common/types"

	"github.com/astaxie/beego/logs"
)

const ACCOUNT_TOKEN_TABLENAME string = "account_token"

var tokenExpire int64 = 2592000000 // 30天,毫秒数

type AccountToken struct {
	Id          int64  `orm:"pk;"`
	AccountId   int64  `orm:"column(account_id)"`
	AccessToken string `orm:"column(access_token)"`
	TokenIp     string `orm:"column(token_ip)"`
	Expires     int64
	Status      int
	Platform    string
	FcmToken    string
	Ctime       int64
	Utime       int64
}

func (r *AccountToken) TableName() string {
	return ACCOUNT_TOKEN_TABLENAME
}

func (r *AccountToken) Using() string {
	return types.OrmDataBaseApi
}

func (r *AccountToken) UsingSlave() string {
	return types.OrmDataBaseApiSlave
}

func GenerateAccountToken(accountId int64, platform string, ip string, fcmToken string) (string, error) {
	bizId, _ := device.GenerateBizId(types.AccessTokenBiz)
	accessToken := tools.Md5(fmt.Sprintf("%dhy0kle@gmail.com%d@%s", bizId, time.Now().UnixNano(), platform))
	var expires int64 = tools.GetUnixMillis() + tokenExpire
	atIns := AccountToken{
		AccountId:   accountId,
		AccessToken: accessToken,
		TokenIp:     ip,
		Expires:     expires,
		Status:      types.StatusValid,
		Platform:    platform,
		FcmToken:    fcmToken,
		Ctime:       tools.GetUnixMillis(),
		Utime:       tools.GetUnixMillis(),
	}

	o := orm.NewOrm()
	o.Using(atIns.Using())
	_, err := o.Insert(&atIns)

	return accessToken, err
}

func GetAccessTokenInfo(token string) (AccountToken, error) {
	var atIns = AccountToken{}
	o := orm.NewOrm()
	o.Using(atIns.Using())
	err := o.QueryTable(atIns.TableName()).Filter("access_token", token).One(&atIns)
	if err != nil && err != orm.ErrNoRows {
		logs.Error("[GetAccessTokenInfo] sql error err:%v", err)
	}

	return atIns, err
}

func UpdateAccessTokenStatusByAccountId(accountId int64, status int) error {
	var atIns = AccountToken{}

	expires := tools.GetUnixMillis()
	if status == types.StatusValid {
		expires = tools.GetUnixMillis() + tokenExpire
	}

	o := orm.NewOrm()
	o.Using(atIns.Using())
	_, err := o.QueryTable(atIns.TableName()).Filter("account_id", accountId).Update(map[string]interface{}{
		"status":  status,
		"expires": expires,
		"utime":   tools.GetUnixMillis(),
	})

	return err
}

// 账户下有效的token集合
func AccountValidToken(accountId int64) (int64, []AccountToken, error) {
	var list []AccountToken
	var m AccountToken

	// 构建查询对象
	qb, _ := orm.NewQueryBuilder(tools.DBDriver())
	qb.Select("*").
		From(m.TableName()).
		Where("account_id = ? AND status = ? AND expires > ?").
		OrderBy("id").
		Desc()

	// 导出 SQL 语句
	sql := qb.String()

	// 执行 SQL 语句
	o := orm.NewOrm()
	o.Using(m.Using())
	num, err := o.Raw(sql, accountId, types.StatusValid, tools.GetUnixMillis()).QueryRows(&list)

	return num, list, err
}

func LatestToken(accountId int64) string {
	// 执行 SQL 语句
	var atIns = AccountToken{}
	o := orm.NewOrm()
	o.Using(atIns.Using())

	err := o.QueryTable(atIns.TableName()).
		Filter("account_id", accountId).
		OrderBy("-id").One(&atIns)

	if err != nil && err != orm.ErrNoRows {
		logs.Error("[LatestToken] db err:%s accountId:%d", err, accountId)
		return ""
	}
	return atIns.AccessToken
}
