package main

import (

	// 数据库初始化

	"fmt"
	"micro-loan/common/dao"
	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"
	"micro-loan/common/models"
	"micro-loan/common/pkg/system/config"
	"micro-loan/common/service"
	"micro-loan/common/tools"
	"micro-loan/common/types"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

var procTitle = "fix-blacklist"

func main() {
	addblacklist()
}

func addblacklist() {

	//获取逾期订单，且逾期天数>=30的 加入黑名单
	overList, _ := GetOverdueOrderList()
	if len(overList) > 0 {
		for _, orderData := range overList {
			repayPlan, _ := models.GetLastRepayPlanByOrderid(orderData.Id)
			_, overdueDays, _ := service.CalculateOverdueLevel(repayPlan.RepayDate)

			// 命中系统黑名单规则，逾期>= 设定天数，触发黑名单事件
			itemName := "overdue_blacklist_day"
			itemValue, _ := config.ValidItemInt(itemName)
			// 逾期>=30 并且 订单状态为 9（逾期）触发黑名单事件
			if overdueDays >= int64(itemValue) && orderData.CheckStatus == types.LoanStatusOverdue {
				accountBase, _ := models.OneAccountBaseByPkId(orderData.UserAccountId)

				if accountBase.Mobile != "" {
					service.AddCustomerRisk(
						accountBase.Id,
						0,
						types.RiskItemMobile,
						types.RiskBlacklist,
						types.RiskReasonHighRisk,
						accountBase.Mobile,
						"overdue>="+tools.Int2Str(itemValue),
						types.RiskReviewPass,
						tools.GetUnixMillis(),
					)
				}
				if accountBase.Identity != "" {
					service.AddCustomerRisk(
						accountBase.Id,
						0,
						types.RiskItemIdentity,
						types.RiskBlacklist,
						types.RiskReasonHighRisk,
						accountBase.Identity,
						"overdue>="+tools.Int2Str(itemValue),
						types.RiskReviewPass,
						tools.GetUnixMillis(),
					)
				}

				logs.Info("[逾期>=30] 加入黑名单 手机 %d ,身份证 %d:", accountBase.Mobile, accountBase.Identity)
			}

		}
	}

	//获取命中规则E013-015的订单
	elist, _ := GetEOrderList()
	if len(elist) > 0 {
		for _, risk := range elist {
			/** E013 同联系人我司申请人当前逾期人数≥3 */

			if risk.HitRegular == "E013" {
				accountBase, _ := models.OneAccountBaseByPkId(risk.AccountId)
				accountProfile, _ := dao.CustomerProfile(risk.AccountId)
				riskCtlE013, _ := config.ValidItemInt64("risk_ctl_E013")
				accountIDs, total, _ := service.SameContactsCustomerOverdueStat(accountProfile.Contact1, accountProfile.Contact2, accountProfile.AccountId)

				if total >= riskCtlE013 {

					//命中E013规则，触发加入黑名单事件，系统自动加入黑名单（命中客户的手机，身份证，联系人手机）

					if accountBase.Mobile != "" {
						service.AddCustomerRisk(
							accountBase.Id, 0, types.RiskItemMobile, types.RiskBlacklist, types.RiskReasonLiar,
							accountBase.Mobile, "E013", types.RiskReviewPass, tools.GetUnixMillis(),
						)
					}
					if accountBase.Identity != "" {
						service.AddCustomerRisk(
							accountBase.Id, 0, types.RiskItemIdentity, types.RiskBlacklist, types.RiskReasonLiar,
							accountBase.Identity, "E013", types.RiskReviewPass, tools.GetUnixMillis(),
						)
					}
					logs.Error("[E013] 加入黑名单 手机 %d ,身份证 %d:", accountBase.Mobile, accountBase.Identity)

					//联系人加入黑名单需要命中公共联系人规则
					commonContact := service.FindCommonContact(accountIDs)
					for _, contact := range commonContact {
						// Trigger(&BlacklistEv{0, types.RiskItemMobile, contact, types.RiskReasonLiar, "E013"})
						if contact != "" {
							service.AddCustomerRisk(
								0, 0, types.RiskItemMobile, types.RiskBlacklist, types.RiskReasonLiar,
								contact, "E013", types.RiskReviewPass, tools.GetUnixMillis(),
							)
						}
						logs.Error("[E013] 公共联系人 手机 %d :", accountBase.Mobile)
					}
					//连带一起命中规则的其他账户
					for _, accountID := range accountIDs {
						accountBase, _ := models.OneAccountBaseByPkId(accountID)
						// Trigger(&BlacklistEv{accountBase.Id, types.RiskItemMobile, accountBase.Mobile, types.RiskReasonLiar, "E013"})
						// Trigger(&BlacklistEv{accountBase.Id, types.RiskItemIdentity, accountBase.Identity, types.RiskReasonLiar, "E013"})
						if accountBase.Mobile != "" {
							service.AddCustomerRisk(
								accountBase.Id, 0, types.RiskItemMobile, types.RiskBlacklist, types.RiskReasonLiar,
								accountBase.Mobile, "E013", types.RiskReviewPass, tools.GetUnixMillis(),
							)
						}
						if accountBase.Identity != "" {
							service.AddCustomerRisk(
								accountBase.Id, 0, types.RiskItemIdentity, types.RiskBlacklist, types.RiskReasonLiar,
								accountBase.Identity, "E013", types.RiskReviewPass, tools.GetUnixMillis(),
							)
						}
						logs.Error("[E013] 连坐== 手机 %d ,身份证 %d:", accountBase.Mobile, accountBase.Identity)
					}

				}
			}
			/** E014 同居住地址的申请人当前逾期人数≥3 */
			if risk.HitRegular == "E014" {
				accountBase, _ := models.OneAccountBaseByPkId(risk.AccountId)
				accountProfile, _ := dao.CustomerProfile(risk.AccountId)

				// fmt.Print(accountProfile)
				// os.Exit(0)
				riskCtlE014, _ := config.ValidItemInt64("risk_ctl_E014")
				accountIDs, total, _ := service.SameResidenceOverdueStat(accountProfile.ResidentCity, accountProfile.ResidentAddress)
				// fmt.Print(total)
				// os.Exit(0)

				if total >= riskCtlE014 {

					//命中E014规则，触发加入黑名单事件，系统自动加入黑名单（命中客户的手机，身份证，居住地址）
					homeAddress := accountProfile.ResidentCity + "," + accountProfile.ResidentAddress
					// Trigger(&BlacklistEv{accountBase.Id, types.RiskItemMobile, accountBase.Mobile, types.RiskReasonLiar, "E014"})
					// Trigger(&BlacklistEv{accountBase.Id, types.RiskItemIdentity, accountBase.Identity, types.RiskReasonLiar, "E014"})
					// Trigger(&BlacklistEv{accountBase.Id, types.RiskItemResidentAddress, homeAddress, types.RiskReasonLiar, "E014"})
					if accountBase.Mobile != "" {
						service.AddCustomerRisk(
							accountBase.Id, 0, types.RiskItemMobile, types.RiskBlacklist, types.RiskReasonLiar,
							accountBase.Mobile, "E014", types.RiskReviewPass, tools.GetUnixMillis(),
						)
					}
					if accountBase.Identity != "" {
						service.AddCustomerRisk(
							accountBase.Id, 0, types.RiskItemIdentity, types.RiskBlacklist, types.RiskReasonLiar,
							accountBase.Identity, "E014", types.RiskReviewPass, tools.GetUnixMillis(),
						)
					}
					if accountProfile.ResidentCity != "" && accountProfile.ResidentAddress != "" {
						service.AddCustomerRisk(
							accountBase.Id, 0, types.RiskItemResidentAddress, types.RiskBlacklist, types.RiskReasonLiar,
							homeAddress, "E014", types.RiskReviewPass, tools.GetUnixMillis(),
						)
					}
					logs.Error("[E014]手机 %d ,身份证 %d,地址 %s:", accountBase.Mobile, accountBase.Identity, homeAddress)

					//连带一起命中规则的其他账户
					for _, accountID := range accountIDs {
						accountBase, _ := models.OneAccountBaseByPkId(accountID)
						// Trigger(&BlacklistEv{accountBase.Id, types.RiskItemMobile, accountBase.Mobile, types.RiskReasonLiar, "E014"})
						// Trigger(&BlacklistEv{accountBase.Id, types.RiskItemIdentity, accountBase.Identity, types.RiskReasonLiar, "E014"})
						if accountBase.Mobile != "" {
							service.AddCustomerRisk(
								accountBase.Id, 0, types.RiskItemMobile, types.RiskBlacklist, types.RiskReasonLiar,
								accountBase.Mobile, "E014", types.RiskReviewPass, tools.GetUnixMillis(),
							)
						}
						if accountBase.Identity != "" {
							service.AddCustomerRisk(
								accountBase.Id, 0, types.RiskItemIdentity, types.RiskBlacklist, types.RiskReasonLiar,
								accountBase.Identity, "E014", types.RiskReviewPass, tools.GetUnixMillis(),
							)
						}
						logs.Error("[E014] 连坐== 手机 %d ,身份证 %d:", accountBase.Mobile, accountBase.Identity)
					}

				}
			}

			/** E015 同单位名称我司申请人当前逾期人数≥3 */
			if risk.HitRegular == "E015" {
				accountBase, _ := models.OneAccountBaseByPkId(risk.AccountId)
				accountProfile, _ := dao.CustomerProfile(risk.AccountId)
				riskCtlE015, _ := config.ValidItemInt64("risk_ctl_E015")
				accountIDs, total, _ := service.SameCompanyOverdueStat(accountProfile.CompanyName)
				if total >= riskCtlE015 {

					//命中E015规则，触发加入黑名单事件，系统自动加入黑名单（命中客户的手机，身份证，单位名称）
					// Trigger(&BlacklistEv{accountBase.Id, types.RiskItemMobile, accountBase.Mobile, types.RiskReasonLiar, "E015"})
					// Trigger(&BlacklistEv{accountBase.Id, types.RiskItemIdentity, accountBase.Identity, types.RiskReasonLiar, "E015"})
					// Trigger(&BlacklistEv{accountBase.Id, types.RiskItemCompany, accountProfile.CompanyName, types.RiskReasonLiar, "E015"})
					if accountBase.Mobile != "" {
						service.AddCustomerRisk(
							accountBase.Id, 0, types.RiskItemMobile, types.RiskBlacklist, types.RiskReasonLiar,
							accountBase.Mobile, "E015", types.RiskReviewPass, tools.GetUnixMillis(),
						)
					}
					if accountBase.Identity != "" {
						service.AddCustomerRisk(
							accountBase.Id, 0, types.RiskItemIdentity, types.RiskBlacklist, types.RiskReasonLiar,
							accountBase.Identity, "E015", types.RiskReviewPass, tools.GetUnixMillis(),
						)
					}
					if accountProfile.CompanyName != "" {
						service.AddCustomerRisk(
							accountBase.Id, 0, types.RiskItemCompany, types.RiskBlacklist, types.RiskReasonLiar,
							accountProfile.CompanyName, "E015", types.RiskReviewPass, tools.GetUnixMillis(),
						)
					}
					logs.Error("[E015] 手机 %d ,身份证 %d ,公司名 %s:", accountBase.Mobile, accountBase.Identity, accountProfile.CompanyName)
					//连带一起命中规则的其他账户
					for _, accountID := range accountIDs {
						accountBase, _ := models.OneAccountBaseByPkId(accountID)
						// Trigger(&BlacklistEv{accountBase.Id, types.RiskItemMobile, accountBase.Mobile, types.RiskReasonLiar, "E015"})
						// Trigger(&BlacklistEv{accountBase.Id, types.RiskItemIdentity, accountBase.Identity, types.RiskReasonLiar, "E015"})
						if accountBase.Mobile != "" {
							service.AddCustomerRisk(
								accountBase.Id, 0, types.RiskItemMobile, types.RiskBlacklist, types.RiskReasonLiar,
								accountBase.Mobile, "E015", types.RiskReviewPass, tools.GetUnixMillis(),
							)
						}
						if accountBase.Identity != "" {
							service.AddCustomerRisk(
								accountBase.Id, 0, types.RiskItemIdentity, types.RiskBlacklist, types.RiskReasonLiar,
								accountBase.Identity, "E015", types.RiskReviewPass, tools.GetUnixMillis(),
							)
						}
						logs.Error("[E015] 连坐== 手机 %d ,身份证 %d:", accountBase.Mobile, accountBase.Identity)
					}
				}
			}
		}
	}
}

// 取给定条数的逾期订单
func GetOverdueOrderList() (list []models.Order, err error) {

	orderM := models.Order{}
	repayPlan := models.RepayPlan{}
	o := orm.NewOrm()
	o.Using(orderM.Using())

	todayNatural := tools.NaturalDay(0)
	sql := fmt.Sprintf(`SELECT o.* FROM %s o
LEFT JOIN %s r ON r.order_id = o.id
WHERE  o.check_status = %d AND r.repay_date < %d`,
		orderM.TableName(),
		repayPlan.TableName(),
		types.LoanStatusOverdue,
		todayNatural,
	)
	_, err = o.Raw(sql).QueryRows(&list)

	return
}

// 获取命中规则E013-015的订单
func GetEOrderList() (list []models.RiskRegularRecord, err error) {

	orderM := models.RiskRegularRecord{}
	o := orm.NewOrm()
	o.Using(orderM.Using())

	sql := fmt.Sprintf(`select * from risk_regular_record where hit_regular in("E013","E014","E015");`)
	_, err = o.Raw(sql).QueryRows(&list)
	return
}
