package dao

import (
	"fmt"
	"micro-loan/common/models"
	"micro-loan/common/tools"
	"micro-loan/common/types"

	"micro-loan/common/lib/redis/storage"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"github.com/garyburd/redigo/redis"
)

func QueryConpon(condStr map[string]interface{}, page, pagesize int) (list []models.Coupon, count int, err error) {
	o := orm.NewOrm()
	m := models.Coupon{}
	o.Using(m.UsingSlave())

	if page < 1 {
		page = 1
	}

	offset := (page - 1) * pagesize
	cond := "1=1"

	if f, ok := condStr["name"]; ok {
		cond = fmt.Sprintf("%s%s'%%%s%%'", cond, " AND name LIKE ", f.(string))
	}
	if f, ok := condStr["status"]; ok {
		cond = fmt.Sprintf("%s%s%d", cond, " AND is_available = ", f.(int))
	}
	if f, ok := condStr["distribute_status"]; ok {
		value, _ := f.(int)
		now := tools.GetUnixMillis()
		if value == 1 {
			cond = fmt.Sprintf("%s%s%d", cond, " AND distribute_start > ", now)
		} else if value == 2 {
			cond = fmt.Sprintf("%s%s%d%s%d%s%d%s", cond, " AND (distribute_start <= ", now, " AND distribute_end >= ", now, " AND is_available = ", types.CouponAvailable, ")")
		} else if value == 3 {
			cond = fmt.Sprintf("%s%s%d", cond, " AND distribute_end < ", now)
		}
	}
	if f, ok := condStr["start_time"]; ok {
		cond = fmt.Sprintf("%s%s%d", cond, " AND distribute_start > ", f.(int64))
	}
	if f, ok := condStr["end_time"]; ok {
		cond = fmt.Sprintf("%s%s%d", cond, " AND distribute_end < ", f.(int64))
	}
	if f, ok := condStr["coupon_type"]; ok {
		cond = fmt.Sprintf("%s%s%d", cond, " AND coupon_type = ", f.(int))
	}
	if f, ok := condStr["distribute_algo"]; ok {
		cond = fmt.Sprintf("%s%s'%s'", cond, " AND distribute_algo = ", f.(string))
	}

	orderBy := "ORDER BY id desc"

	limit := fmt.Sprintf("LIMIT %d, %d", offset, pagesize)

	sql := "SELECT * FROM coupon WHERE " + cond
	sqlData := fmt.Sprintf("%s %s %s", sql, orderBy, limit)

	sqlCount := "SELECT count(id) FROM coupon WHERE " + cond

	r := o.Raw(sqlData)
	_, err = r.QueryRows(&list)

	r = o.Raw(sqlCount)
	r.QueryRow(&count)

	return
}

type AccountCouponInfo struct {
	Id            int64
	UserAccountId int64
	OrderId       int64
	Name          string
	CouponType    types.CouponType
	Status        types.CouponStatus
	Ctime         int64
	UsedTime      int64
	ExpireDate    int64
	Amount        int64
}

func QueryAccountCoupon(condStr map[string]interface{}, page, pagesize int) (list []AccountCouponInfo, count int, err error) {
	o := orm.NewOrm()
	c := models.Coupon{}
	a := models.AccountCoupon{}
	o.Using(c.UsingSlave())

	if page < 1 {
		page = 1
	}

	offset := (page - 1) * pagesize
	cond := "1=1"

	if f, ok := condStr["coupon_type"]; ok {
		cond = fmt.Sprintf("%s%s%d", cond, " AND c.coupon_type = ", f.(int))
	}
	if f, ok := condStr["coupon_status"]; ok {
		cond = fmt.Sprintf("%s%s%d", cond, " AND a.status = ", f.(int))
	}
	if f, ok := condStr["name"]; ok {
		cond = fmt.Sprintf("%s%s'%%%s%%'", cond, " AND c.name LIKE ", f.(string))
	}
	if f, ok := condStr["distr_algo"]; ok {
		cond = fmt.Sprintf("%s%s'%s'", cond, " AND c.distribute_algo = ", f.(string))
	}
	if f, ok := condStr["account_id"]; ok {
		cond = fmt.Sprintf("%s%s%d", cond, " AND a.user_account_id = ", f.(int64))
	}
	if f, ok := condStr["coupon_id"]; ok {
		cond = fmt.Sprintf("%s%s%d", cond, " AND a.coupon_id = ", f.(int64))
	}
	if f, ok := condStr["distr_range_start"]; ok {
		cond = fmt.Sprintf("%s%s%d", cond, " AND a.ctime > ", f.(int64))
	}
	if f, ok := condStr["distr_range_end"]; ok {
		cond = fmt.Sprintf("%s%s%d", cond, " AND a.ctime < ", f.(int64))
	}
	if f, ok := condStr["used_range_start"]; ok {
		cond = fmt.Sprintf("%s%s%d", cond, " AND a.used_time > ", f.(int64))
	}
	if f, ok := condStr["used_range_end"]; ok {
		cond = fmt.Sprintf("%s%s%d", cond, " AND a.used_time < ", f.(int64))
	}

	sql := fmt.Sprintf(`SELECT c.coupon_type, c.name, a.id, a.user_account_id, a.order_id, a.status, a.ctime, a.used_time, a.expire_date, a.amount
FROM %s a LEFT JOIN %s c ON a.coupon_id = c.id WHERE `,
		a.TableName(), c.TableName()) + cond

	sqlCount := fmt.Sprintf(`SELECT count(a.order_id) 
FROM %s a LEFT JOIN %s c ON a.coupon_id = c.id WHERE `,
		a.TableName(), c.TableName()) + cond

	orderBy := "ORDER BY a.id"

	limit := fmt.Sprintf("LIMIT %d, %d", offset, pagesize)

	sqlData := fmt.Sprintf("%s %s %s", sql, orderBy, limit)

	r := o.Raw(sqlData)
	_, err = r.QueryRows(&list)

	r = o.Raw(sqlCount)
	r.QueryRow(&count)

	return
}

func GetCouponById(id int64) (one models.Coupon, err error) {
	o := orm.NewOrm()
	o.Using(one.Using())

	err = o.QueryTable(one.TableName()).
		Filter("id", id).
		One(&one)

	return
}

func GetCouponByUserType(name string) (list []models.Coupon, err error) {
	m := models.Coupon{}
	o := orm.NewOrm()
	o.Using(m.UsingSlave())

	_, err = o.QueryTable(m.TableName()).
		Filter("distribute_algo", name).
		All(&list)

	return
}

func UpdateCoupon(one *models.Coupon) error {
	one.Utime = tools.GetUnixMillis()

	o := orm.NewOrm()
	o.Using(one.Using())

	_, err := o.Update(one)
	if err != nil {
		logs.Error("[UpdateCoupon] update error id:%d, err:%v", one.Id, err)
	}

	return err
}

func GetAccountCouponById(id int64) (one models.AccountCoupon, err error) {
	o := orm.NewOrm()
	o.Using(one.Using())

	err = o.QueryTable(one.TableName()).
		Filter("id", id).
		One(&one)

	return
}

func UpdateAccountCoupon(one *models.AccountCoupon) error {
	one.Utime = tools.GetUnixMillis()

	o := orm.NewOrm()
	o.Using(one.Using())

	_, err := o.Update(one)
	if err != nil {
		logs.Error("[UpdateAccountCoupon] update error id:%d, err:%v", one.Id, err)
	}

	return err
}

func GetAccountFrozenCouponByOrder(accountId, orderId int64) (one models.AccountCoupon, err error) {
	o := orm.NewOrm()
	o.Using(one.Using())

	err = o.QueryTable(one.TableName()).
		Filter("user_account_id", accountId).
		Filter("order_id", orderId).
		Filter("status", types.CouponStatusFrozen).
		One(&one)

	return
}

func GetAccountCouponByOrderAndStatus(accountId, orderId int64, status int) (one models.AccountCoupon, err error) {
	o := orm.NewOrm()
	o.Using(one.Using())

	err = o.QueryTable(one.TableName()).
		Filter("user_account_id", accountId).
		Filter("order_id", orderId).
		Filter("status", status).
		One(&one)
	return
}

func GetAvailableAccountCoupon(couponId int64, limit int) (list []models.AccountCoupon, err error) {
	m := models.AccountCoupon{}

	o := orm.NewOrm()
	o.Using(m.Using())

	sql := fmt.Sprintf(`SELECT * FROM %s
WHERE coupon_id = %d AND status = %d limit %d`,
		m.TableName(),
		couponId,
		types.CouponStatusAvailable,
		limit,
	)

	_, err = o.Raw(sql).QueryRows(&list)

	return
}

func AddCoupon(coupon *models.Coupon) error {
	coupon.Ctime = tools.GetUnixMillis()

	o := orm.NewOrm()
	o.Using(coupon.Using())
	_, err := o.Insert(coupon)
	if err != nil {
		logs.Error("[AddCoupon] Insert model failed err:%v", err)
	}
	return err
}

func AddAccountCoupon(accountCoupon *models.AccountCoupon) (int64, error) {
	o := orm.NewOrm()
	o.Using(accountCoupon.Using())
	id, err := o.Insert(accountCoupon)
	if err != nil {
		logs.Error("[AddAccountCoupon] Insert model failed err:%v", err)
	}
	return id, err
}

func QueryExpireAccountCoupon(timestamp int64) (list []models.AccountCoupon, err error) {
	o := orm.NewOrm()
	m := models.AccountCoupon{}
	o.Using(m.UsingSlave())

	sql := fmt.Sprintf(`SELECT * FROM %s
WHERE status = %d AND valid_end <= %d`,
		m.TableName(),
		types.CouponStatusAvailable,
		timestamp)
	_, err = o.Raw(sql).QueryRows(&list)

	return
}

func QueryRejectOrderCoupon(startDate, endDate int64, limit int) (list []models.AccountCoupon, err error) {
	o := orm.NewOrm()
	c := models.AccountCoupon{}
	order := models.Order{}
	o.Using(c.UsingSlave())

	sql := fmt.Sprintf(`SELECT c.* FROM %s c
LEFT JOIN %s o ON c.order_id = o.id 
WHERE c.status = %d AND o.check_status = %d AND o.check_time >= %d AND o.check_time < %d 
limit %d`,
		c.TableName(), order.TableName(),
		types.CouponStatusFrozen,
		types.LoanStatusReject,
		startDate, endDate,
		limit)
	_, err = o.Raw(sql).QueryRows(&list)

	return
}

type ApiCouponInfo struct {
	Id             int64
	CouponType     types.CouponType
	DiscountRate   int64
	DiscountDay    int64
	DiscountAmount int64
	ValidStart     int64
	ValidDate      int64
	ValidMin       int64
	DiscountMax    int64
	ValidAmount    int64
	Amount         int64
	Loan           int64
	IsAvaliable    int
}

type CouponInfoList []ApiCouponInfo

func (s CouponInfoList) Len() int {
	return len(s)
}

func (s CouponInfoList) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s CouponInfoList) Less(i, j int) bool {
	if s[i].IsAvaliable != s[j].IsAvaliable {
		return s[i].IsAvaliable < s[j].IsAvaliable
	}

	if s[i].CouponType == types.CouponTypeLimit && s[j].CouponType == types.CouponTypeLimit {
		return s[i].Id > s[j].Id
	} else if s[i].CouponType == types.CouponTypeLimit {
		return true
	} else if s[j].CouponType == types.CouponTypeLimit {
		return false
	} else {
		return s[i].Id > s[j].Id
	}
}

func QueryAccountCouponList(accountId int64, couponTypes []types.CouponType, pagesize, offset int, expireDay int64) (list CouponInfoList, err error) {
	o := orm.NewOrm()
	c := models.Coupon{}
	a := models.AccountCoupon{}
	o.Using(c.UsingSlave())

	nowDate := tools.GetUnixMillis()
	expireDate := tools.NaturalDay(expireDay * -1)
	tmpList := make([]int, 0)
	for _, v := range couponTypes {
		tmpList = append(tmpList, int(v))
	}

	sql := fmt.Sprintf(`SELECT c.coupon_type, c.discount_rate, c.discount_day, c.discount_amount, c.valid_min, c.discount_max, a.id, a.valid_end as valid_date, a.valid_start, a.status as is_avaliable, if(c.coupon_type = %d,0,1) as type_order
FROM %s c LEFT JOIN %s a ON a.coupon_id = c.id 
WHERE a.user_account_id = %d AND ((a.status = %d AND a.valid_end > %d) OR (a.status = %d AND a.expire_date > %d)) AND c.coupon_type in (%s)
ORDER BY a.status asc, type_order asc, a.ctime desc
LIMIT %d, %d `,
		types.CouponTypeLimit,
		c.TableName(), a.TableName(),
		accountId,
		types.CouponStatusAvailable, nowDate,
		types.CouponStatusInvalid, expireDate,
		tools.ArrayToString(tmpList, ","),
		offset, pagesize)

	r := o.Raw(sql)
	_, err = r.QueryRows(&list)

	return
}

func QueryAccountCouponActive(accountId int64, couponTypes []types.CouponType) (list CouponInfoList, err error) {
	o := orm.NewOrm()
	c := models.Coupon{}
	a := models.AccountCoupon{}
	o.Using(c.UsingSlave())

	tmpList := make([]int, 0)
	for _, v := range couponTypes {
		tmpList = append(tmpList, int(v))
	}

	nowDate := tools.GetUnixMillis()

	sql := fmt.Sprintf(`SELECT c.coupon_type, c.discount_rate, c.discount_day, c.discount_amount, c.valid_min, c.discount_max, a.id, a.valid_end as valid_date, a.valid_start
FROM %s c LEFT JOIN %s a ON a.coupon_id = c.id 
WHERE a.user_account_id = %d AND a.status = %d AND a.valid_end > %d AND c.coupon_type in (%s)
ORDER BY a.ctime desc`,
		c.TableName(), a.TableName(),
		accountId,
		types.CouponStatusAvailable, nowDate,
		tools.ArrayToString(tmpList, ","))

	r := o.Raw(sql)
	_, err = r.QueryRows(&list)

	return
}

func GetHistoryCoupon() []string {
	o := orm.NewOrm()
	c := models.Coupon{}

	o.Using(c.UsingSlave())

	sql := fmt.Sprintf(`SELECT DISTINCT(distribute_algo) 
FROM %s
WHERE distribute_algo != "";`,
		c.TableName())

	list := make([]string, 0)
	r := o.Raw(sql)
	r.QueryRows(&list)

	return list
}

func GetAccountNewCoupon(accountId int64) (list []models.AccountCoupon, err error) {
	m := models.AccountCoupon{}
	o := orm.NewOrm()
	o.Using(m.Using())

	_, err = o.QueryTable(m.TableName()).
		Filter("user_account_id", accountId).
		Filter("is_new", types.CouponUnread).
		All(&list)

	return
}

func GetAccountTask(accountId int64, taskType types.AccountTask) (one models.AccountTask, err error) {
	o := orm.NewOrm()
	o.Using(one.Using())

	err = o.QueryTable(one.TableName()).
		Filter("account_id", accountId).
		Filter("task_type", taskType).
		One(&one)

	return
}

func UpdateAccountTask(one *models.AccountTask) error {
	one.Utime = tools.GetUnixMillis()

	o := orm.NewOrm()
	o.Using(one.Using())

	_, err := o.Update(one)
	if err != nil {
		logs.Error("[UpdateAccountTask] update error id:%d, err:%v", one.Id, err)
	}

	return err
}

func AddAccountTask(one *models.AccountTask) error {
	one.Ctime = tools.GetUnixMillis()

	o := orm.NewOrm()
	o.Using(one.Using())
	_, err := o.Insert(one)
	if err != nil {
		logs.Error("[AddAccountTask] Insert model failed err:%v", err)
	}
	return err
}

func QueryAccountTaskByInviter(accountId int64) (list []models.AccountTask, err error) {
	m := models.AccountTask{}
	o := orm.NewOrm()
	o.Using(m.Using())

	_, err = o.QueryTable(m.TableName()).
		Filter("inviter_id", accountId).
		OrderBy("account_id", "task_type").
		All(&list)

	return
}

func GetCouponTotalNumInfo(id int64) int64 {
	totalKey := beego.AppConfig.String("coupon_total")

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	num, _ := redis.Int64(storageClient.Do("HGET", totalKey, id))

	return num
}
