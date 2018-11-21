package main

import (

	// 数据库初始化

	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"
	"micro-loan/common/service"

	"github.com/astaxie/beego/logs"
)

var procTitle = "customer-tags"

func main() {
	customerTags()
}

func customerTags() {

	customerList, _ := service.CustomerWaitingTag()
	for _, customer := range customerList {
		// 写队列
		handleCustomerTags(customer.Id)
	}

}
func handleCustomerTags(accountID int64) {

	//获取用户全部订单
	tags := service.CustomerTags(accountID)
	//doCustomerTags 用户激活后为用户打标签 ，2：目标客户
	service.UpdateCustomer(accountID, tags)
	logs.Info("[customerTagsQueue]  已更新, tag :", tags, ", accountID:", accountID)
}
