package controllers

import (
	"encoding/json"
	"micro-loan/common/pkg/rbac"
	"micro-loan/common/pkg/ticket/performance"
	"micro-loan/common/types"
)

//"micro-loan/common/thirdparty/sms"

type IndexController struct {
	BaseController
}

func (c *IndexController) Prepare() {
	// 调用上一级的 Prepare 方法
	c.BaseController.Prepare()

	c.Data["Controller"] = "index"
}

func (c *IndexController) Get() {
	c.Data["Website"] = "beego.me"
	c.Data["Email"] = "astaxie@gmail.com"
	c.Data["Action"] = "index"

	c.Layout = "layout.html"
	c.TplName = "index.html"

	//sms.SendSms("8618911039591", "jcfirstmsg");
}

func (c *IndexController) Dashboard() {
	testToken := c.GetString("test")

	if c.RoleType == types.RoleTypeUrge || c.RoleType == types.RoleTypeRepayReminder || testToken == "test" {
		uid, _ := c.GetInt64("uid")
		if uid == 0 {
			uid = c.AdminUid
		}
		roleID, _ := c.GetInt64("rid")
		if roleID == 0 {
			roleID = c.RoleID
		}

		lastestStatsData, err := performance.GetDailyLatestStats(uid)
		if err != nil {
			c.Layout = "layout.html"
			c.TplName = "index.html"
			return
		}

		itemID := lastestStatsData.TicketItemID
		hour := lastestStatsData.Hour
		if rbac.GetRoleLevel(roleID) == types.RoleLeader && (c.RoleType != types.RoleTypeRepayReminder || testToken == "test") {
			lastestGroupStats := performance.GetGroupDailyLatestStats(roleID, hour, itemID)
			c.Data["lastestGroupStats"] = lastestGroupStats
			c.Data["groupLeaderRoleID"] = roleID
			c.Data["groupTargetRepayRate"] = performance.GetUrgeGroupTargetRepayRate(itemID)
			c.Data["groupRankingList"] = performance.GetGroupListStats(roleID, hour, itemID)
			jsonGroupChartData, _ := json.Marshal(performance.GetSingleGroupChart(roleID, hour, itemID))
			c.Data["jsonGroupChartData"] = string(jsonGroupChartData)
			c.TplName = "dashboard/urge_leader.html"
			//
		} else {
			c.TplName = "dashboard/urge_employee.html"
		}
		{
			// 员工dashboard 与 leader dashboard 中的个人情况
			rankingList := performance.GetCurrentRanking(lastestStatsData.TicketItemID, lastestStatsData.Hour)
			processChartDatas := performance.GetPersonalStatsList(uid, lastestStatsData.TicketItemID)
			jsonProcessChartDatas, _ := json.Marshal(processChartDatas)
			c.Data["rankingList"] = rankingList
			c.Data["jsonProcessChartDatas"] = string(jsonProcessChartDatas)
		}

		c.Data["lastestStatsData"] = lastestStatsData
		c.Data["standardRepayRate"] = lastestStatsData.TargetRepayRate

		c.Layout = "layout.html"

		c.LayoutSections = make(map[string]string)
		c.LayoutSections["CssPlugin"] = "plugin/css.html"
		c.LayoutSections["JsPlugin"] = "plugin/js.html"
		c.LayoutSections["Scripts"] = "dashboard/urge_employee_scripts.html"
		return
	}
	c.Layout = "layout.html"
	c.TplName = "index.html"
}
