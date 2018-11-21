package service

import (
	"fmt"
	"sort"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"

	"micro-loan/common/dao"
	"micro-loan/common/lib/redis/storage"
	"micro-loan/common/models"
	"micro-loan/common/pkg/repayplan"
	"micro-loan/common/tools"
	"micro-loan/common/types"

	"micro-loan/common/pkg/coupon_event"

	"github.com/astaxie/beego"
	"github.com/garyburd/redigo/redis"
)

var MinDiscountStr = "loan < DiscountMin, %d, %d"
var NilDiscountStr = "unexcept coupon type, %d"

func GetAllCoupon(condStr map[string]interface{}, page, pagesize int) (list []models.Coupon, count int, err error) {
	list, count, err = dao.QueryConpon(condStr, page, pagesize)
	if err != nil {
		return
	}

	totalKey := beego.AppConfig.String("coupon_total")
	amountKey := beego.AppConfig.String("coupon_amount")
	usedKey := beego.AppConfig.String("coupon_used")

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	for i := 0; i < len(list); i++ {
		list[i].DistributeAll, _ = redis.Int64(storageClient.Do("HGET", totalKey, list[i].Id))
		list[i].UsedNum, _ = redis.Int64(storageClient.Do("HGET", usedKey, list[i].Id))
		list[i].UsedAmount, _ = redis.Int64(storageClient.Do("HGET", amountKey, list[i].Id))
	}

	return
}

func GetAllAccountCoupon(condStr map[string]interface{}, page, pagesize int) (list []dao.AccountCouponInfo, count int, err error) {
	list, count, err = dao.QueryAccountCoupon(condStr, page, pagesize)
	for i, _ := range list {
		if list[i].Status == types.CouponStatusFrozen {
			list[i].Amount = 0
		}
	}

	return
}

func CalcCouponAmount(loan, amount int64, period int, coupon *dao.ApiCouponInfo, product *models.Product) (int64, int64, error) {
	switch coupon.CouponType {
	case types.CouponTypeRedPacket:
		{
			if amount < coupon.ValidMin {
				return 0, 0, fmt.Errorf("amount < min valid %d, %d", amount, coupon.ValidMin)
			}

			discountAmount := coupon.DiscountAmount

			if discountAmount > coupon.DiscountMax {
				discountAmount = coupon.DiscountMax
			}

			return amount, discountAmount, nil
		}
	case types.CouponTypeDiscount:
		{
			if amount < coupon.ValidMin {
				return 0, 0, fmt.Errorf("amount < min valid %d, %d", amount, coupon.ValidMin)
			}

			discountAmount := amount - amount*coupon.DiscountRate/100

			if discountAmount > coupon.DiscountMax {
				discountAmount = coupon.DiscountMax
			}

			return amount, discountAmount, nil
		}

	case types.CouponTypeInterest:
		{
			if amount < coupon.ValidMin {
				return 0, 0, fmt.Errorf("amount < min valid %d, %d", amount, coupon.ValidMin)
			}

			_, interest, _ := repayplan.CalcRepayInfoV2(loan, *product, period)
			discountAmount := interest * coupon.DiscountDay / int64(period)

			if discountAmount > coupon.DiscountMax {
				discountAmount = coupon.DiscountMax
			}

			return amount, discountAmount, nil
		}
	case types.CouponTypeLimit:
		{
			if loan+coupon.DiscountAmount > coupon.DiscountMax {
				return 0, 0, fmt.Errorf("loan > max valid %d > %d", loan+coupon.DiscountAmount, coupon.DiscountMax)
			}

			total, _, _ := repayplan.CalcRepayInfoV2(loan+coupon.DiscountAmount, *product, period)

			discountAmount := coupon.DiscountAmount

			return total, discountAmount, nil
		}
	}

	err := fmt.Errorf(NilDiscountStr, coupon.CouponType)
	return 0, 0, err
}

func OrderUseCoupon(loan, amount int64, period int, order *models.Order, accountCouponId int64) (int64, int64, types.CouponType, error) {
	accountCoupon, err := dao.GetAccountCouponById(accountCouponId)
	if err != nil {
		logs.Error("[OrderUseCoupon] GetAccountCouponById has wrong couponId:%d, orderId:%d, err:%v", accountCouponId, order.Id, err)
		return 0, 0, types.CouponTypeRedPacket, err
	}

	product, err := models.GetProduct(order.ProductId)
	if err != nil {
		logs.Error("[OrderUseCoupon] GetProduct has wrong couponId:%d, orderId:%d, err:%v", accountCouponId, order.Id, err)
		return 0, 0, types.CouponTypeRedPacket, err
	}

	if accountCoupon.Status != types.CouponStatusAvailable {
		logs.Error("[OrderUseCoupon] account coupon status has wrong orderId:%d, couponId:%d, status:%d, err:%v", order.Id, accountCouponId, accountCoupon.Status, err)
		return 0, 0, types.CouponTypeRedPacket, fmt.Errorf("coupon status wrong")
	}

	coupon, err := dao.GetCouponById(accountCoupon.CouponId)
	if err != nil {
		logs.Warn("[OrderUseCoupon] GetCouponById wrong orderId:%d, couponId:%d, err:%v", order.Id, accountCoupon.CouponId, err)
		MakeAccountCouponInvalid(&accountCoupon)
		return 0, 0, types.CouponTypeRedPacket, err
	}

	nowDate := tools.GetUnixMillis()
	if nowDate > accountCoupon.ValidEnd {
		logs.Warn("[OrderUseCoupon] conpon is expire orderId:%d, couponId:%d, nowDate:%d, endDate:%d, err:%v", order.Id, accountCoupon.CouponId, nowDate, accountCoupon.ValidEnd, err)
		MakeAccountCouponInvalid(&accountCoupon)
		return 0, 0, types.CouponTypeRedPacket, fmt.Errorf("coupon invalid")
	}

	if nowDate < accountCoupon.ValidStart {
		logs.Warn("[OrderUseCoupon] conpon is not ready orderId:%d, couponId:%d, nowDate:%d, startDate:%d, err:%v", order.Id, accountCoupon.CouponId, nowDate, accountCoupon.ValidStart, err)
		return 0, 0, types.CouponTypeRedPacket, fmt.Errorf("coupon invalid")
	}

	apiCoupon := dao.ApiCouponInfo{}
	apiCoupon.CouponType = coupon.CouponType
	apiCoupon.DiscountAmount = coupon.DiscountAmount
	apiCoupon.DiscountRate = coupon.DiscountRate
	apiCoupon.DiscountDay = coupon.DiscountDay
	apiCoupon.DiscountMax = coupon.DiscountMax

	newAmount, disamount, err := CalcCouponAmount(loan, amount, period, &apiCoupon, &product)
	if err != nil {
		logs.Warn("[CalcCouponAmount] conpon amount invalid orderId:%d, couponId:%d, loan:%d, amount:%d, err:%v", order.Id, accountCoupon.CouponId, loan, amount, err)

		MakeAccountCouponAvailable(&accountCoupon)
		return 0, 0, types.CouponTypeRedPacket, err
	}

	if accountCoupon.UsedTime == 0 {
		coupon_event.IncrCouponDailyUsed(accountCoupon.CouponId, 1)
	}

	accountCoupon.Amount = disamount
	accountCoupon.OrderId = order.Id
	accountCoupon.UsedTime = tools.GetUnixMillis()
	accountCoupon.Status = types.CouponStatusFrozen
	dao.UpdateAccountCoupon(&accountCoupon)

	if coupon.CouponType == types.CouponTypeLimit {
		logs.Info("[OrderUseCoupon] conpon is limit type skip orderId:%d, couponId:%d, err:%v", order.Id, accountCoupon.CouponId, err)

		return newAmount, disamount + loan, coupon.CouponType, nil
	} else {
		return amount, loan, coupon.CouponType, nil
	}
}

func HandleCoupon(order *models.Order) (models.AccountCoupon, error) {
	accountCoupon, err := dao.GetAccountFrozenCouponByOrder(order.UserAccountId, order.Id)
	if err != nil {
		if err != orm.ErrNoRows {
			logs.Warn("[HandleCoupon] GetAccountCouponByOrder wrong orderId:%d, err:%v", order.Id, err)
		}

		return accountCoupon, err
	}

	if accountCoupon.Status != types.CouponStatusFrozen {
		err := fmt.Errorf("Coupon status wrong %d", accountCoupon.Status)
		return accountCoupon, err
	}

	accountCoupon.Status = types.CouponStatusUsed
	accountCoupon.EffectiveDate = tools.GetUnixMillis()
	dao.UpdateAccountCoupon(&accountCoupon)

	amountKey := beego.AppConfig.String("coupon_amount")
	usedKey := beego.AppConfig.String("coupon_used")

	storageClient := storage.RedisStorageClient.Get()
	defer storageClient.Close()

	storageClient.Do("HINCRBY", usedKey, accountCoupon.CouponId, 1)
	storageClient.Do("HINCRBY", amountKey, accountCoupon.CouponId, accountCoupon.Amount)

	coupon_event.IncrCouponDailySucc(accountCoupon.CouponId, 1)

	return accountCoupon, nil
}

func MakeAccountCouponInvalid(accountCoupon *models.AccountCoupon) {
	accountCoupon.Status = types.CouponStatusInvalid
	accountCoupon.ExpireDate = tools.GetUnixMillis()
	dao.UpdateAccountCoupon(accountCoupon)
}

func MakeAccountCouponAvailable(accountCoupon *models.AccountCoupon) {
	if accountCoupon.Status != types.CouponStatusFrozen {
		return
	}

	accountCoupon.OrderId = 0
	accountCoupon.Amount = 0
	accountCoupon.Status = types.CouponStatusAvailable

	dao.UpdateAccountCoupon(accountCoupon)
}

func AddCoupon(coupon *models.Coupon) error {
	return dao.AddCoupon(coupon)
}

func HandleCouponEvent(trigger coupon_event.CouponEventTrigger, data interface{}) {
	coupon_event.StartCouponEvent(trigger, data)
}

func CheckCouponListRequired(parameter map[string]interface{}) bool {
	requiredParameter := map[string]bool{
		"offset": true,
	}

	return tools.CheckRequiredParameter(parameter, requiredParameter)
}

func CheckCouponActiveRequired(parameter map[string]interface{}) bool {
	requiredParameter := map[string]bool{
		"period": true,
		"loan":   true,
		"amount": true,
	}

	return tools.CheckRequiredParameter(parameter, requiredParameter)
}

func CheckCouponNewRequired(parameter map[string]interface{}) bool {
	requiredParameter := map[string]bool{}

	return tools.CheckRequiredParameter(parameter, requiredParameter)
}

func CheckMarkNewRequired(parameter map[string]interface{}) bool {
	requiredParameter := map[string]bool{}

	return tools.CheckRequiredParameter(parameter, requiredParameter)
}

func QueryAccountCoupon(accountId int64, couponTypes []types.CouponType, page int, data map[string]interface{}) {
	pagesize := 10
	offset := page * pagesize

	list, err := dao.QueryAccountCouponList(accountId, couponTypes, pagesize, offset, 15)
	if err != nil {
		logs.Error("[QueryAccountCoupon] QueryAccountCouponList error accountId:%d, err:%v", accountId, err)
	}

	if len(list) == 0 {
		data["size"] = 0
		data["offset"] = 0
		data["data"] = make([]string, 0)
		return
	}

	data["size"] = len(list)
	data["offset"] = page + 1

	var dataSet [](map[string]interface{})
	for _, l := range list {
		subSet := map[string]interface{}{
			"id":              l.Id,
			"coupon_type":     l.CouponType,
			"discount_rate":   l.DiscountRate,
			"discount_day":    l.DiscountDay,
			"discount_amount": l.DiscountAmount,
			"valid_date":      l.ValidDate,
			"valid_start":     l.ValidStart,
			"valid_min":       l.ValidMin,
			"discount_max":    l.DiscountMax,
			"is_avaliable":    l.IsAvaliable,
		}

		dataSet = append(dataSet, subSet)
	}

	data["data"] = dataSet
}

func QueryAccountCouponActive(accountId int64, couponTypes []types.CouponType, loan int64, amount int64, period int, dataProduct *models.Product, data map[string]interface{}) {
	emptyList := make([]string, 0)
	hasAvaliable := 0

	list, err := dao.QueryAccountCouponActive(accountId, couponTypes)
	if err != nil {
		logs.Error("[QueryAccountCouponActive] QueryAccountCouponActive error accountId:%d, err:%v", accountId, err)
	}

	if len(list) == 0 {
		data["size"] = 0
		data["data"] = emptyList
		data["is_avaliable"] = hasAvaliable
		return
	}

	nowDate := tools.GetUnixMillis()

	data["size"] = len(list)
	for i, _ := range list {
		if amount < list[i].ValidMin {
			list[i].IsAvaliable = int(types.CouponStatusInvalid)
		} else if list[i].ValidStart > nowDate {
			list[i].IsAvaliable = int(types.CouponStatusInvalid)
		} else {
			newAmount, discount, err := CalcCouponAmount(loan, amount, period, &list[i], dataProduct)
			if err != nil {
				list[i].IsAvaliable = int(types.CouponStatusInvalid)
			} else {
				if list[i].CouponType == types.CouponTypeLimit {
					list[i].ValidAmount = discount
					list[i].Amount = newAmount
					list[i].Loan = loan + discount
				} else {
					list[i].ValidAmount = discount
					list[i].Amount = amount - discount
					list[i].Loan = loan
				}
				list[i].IsAvaliable = int(types.CouponStatusAvailable)
				hasAvaliable = 1
			}
		}
	}

	sort.Sort(list)

	var dataSet [](map[string]interface{})
	for _, l := range list {
		subSet := map[string]interface{}{
			"id":              l.Id,
			"coupon_type":     l.CouponType,
			"discount_rate":   l.DiscountRate,
			"discount_day":    l.DiscountDay,
			"discount_amount": l.DiscountAmount,
			"valid_amount":    l.ValidAmount,
			"amount":          l.Amount,
			"loan":            l.Loan,
			"valid_date":      l.ValidDate,
			"valid_start":     l.ValidStart,
			"valid_min":       l.ValidMin,
			"discount_max":    l.DiscountMax,
			"is_avaliable":    l.IsAvaliable,
		}

		dataSet = append(dataSet, subSet)
	}

	data["data"] = dataSet
	data["is_avaliable"] = hasAvaliable
}

func ActiveCoupon(id int64) error {
	m, err := dao.GetCouponById(id)
	if err != nil {
		return err
	}

	if m.IsAvailable == types.CouponAvailable {
		m.IsAvailable = types.CouponInvalid
		return dao.UpdateCoupon(&m)
	}

	if tools.GetUnixMillis() > m.DistributeEnd {
		return fmt.Errorf("优惠券已经过了派发时间,暂不能重新启用,如果需要再次启用,请重新编辑优惠券信息")
	}

	num := dao.GetCouponTotalNumInfo(m.Id)
	if m.DistributeSize != 0 && num >= m.DistributeSize {
		return fmt.Errorf("优惠券已经没有可派发的数量,暂不能重新启用,如果需要再次启用,请重新编辑优惠券信息")
	}

	m.IsAvailable = types.CouponAvailable
	return dao.UpdateCoupon(&m)
}

func ModifyCoupon(m *models.Coupon, status int, count int64, endDate int64) error {
	nowStr := tools.MDateMHSDate(tools.GetUnixMillis())
	nowZero := tools.GetDateParseBackend(nowStr)*1000 + 3600*24*1000 - 1000

	if m.DistributeEnd != endDate && endDate <= nowZero {
		return fmt.Errorf("终止日期不能小于当天")
	}

	num := dao.GetCouponTotalNumInfo(m.Id)
	if count != 0 && num >= count {
		return fmt.Errorf("发放数量不能小于已发放数量")
	}

	m.DistributeEnd = endDate
	m.DistributeSize = count
	m.IsAvailable = status
	return dao.UpdateCoupon(m)
}

func GetHistoryCoupon() []string {
	return dao.GetHistoryCoupon()
}

func QueryNewAccountCoupon(accountId int64) (list []models.AccountCoupon, err error) {
	list, err = dao.GetAccountNewCoupon(accountId)

	return
}

func MarkNewAccountCoupon(accountId int64) {
	list, _ := dao.GetAccountNewCoupon(accountId)

	for _, v := range list {
		v.IsNew = types.CouponRead

		dao.UpdateAccountCoupon(&v)
	}
}

func QueryCouponRecord(id int64, page, pagesize int) (list []models.CouponDetail, count int64, err error) {
	o := orm.NewOrm()
	c := models.CouponDetail{}
	o.Using(c.UsingSlave())

	if page < 1 {
		page = 1
	}

	offset := (page - 1) * pagesize
	if page == 1 {
		pagesize--
	} else {
		offset--
	}

	sql := fmt.Sprintf(`SELECT *
FROM %s WHERE coupon_id = %d`,
		c.TableName(), id)

	orderBy := "ORDER BY coupon_date desc"

	limit := fmt.Sprintf("LIMIT %d, %d", offset, pagesize)

	sqlData := fmt.Sprintf("%s %s %s", sql, orderBy, limit)

	r := o.Raw(sqlData)
	_, err = r.QueryRows(&list)

	r = o.Raw(sqlData)
	r.QueryRow(&count)

	if page == 1 {
		storageClient := storage.RedisStorageClient.Get()
		defer storageClient.Close()

		nowStr := tools.MDateMHSDate(tools.GetUnixMillis())
		key := fmt.Sprintf("coupon:%d:%s", id, nowStr)

		totalNum, _ := redis.Int(storageClient.Do("HGET", key, coupon_event.CouponKeyTotal))
		succNum, _ := redis.Int(storageClient.Do("HGET", key, coupon_event.CouponKeySucc))
		usedNum, _ := redis.Int(storageClient.Do("HGET", key, coupon_event.CouponKeyUsed))

		count := int(dao.GetCouponTotalNumInfo(id))

		record := models.CouponDetail{}
		record.CouponId = id
		record.UsedNum = usedNum
		record.TotalNum = totalNum
		record.SuccNum = succNum
		record.CouponDate = tools.GetUnixMillis()
		record.Ctime = tools.GetUnixMillis()
		if count > 0 {
			record.SuccRate = record.SuccNum * 100 / count
			record.UsedRate = record.UsedNum * 100 / count
		}

		newList := make([]models.CouponDetail, 0)
		newList = append(newList, record)
		newList = append(newList, list...)
		list = newList

		count++
	}

	return
}
