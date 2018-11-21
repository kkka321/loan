package models

import "micro-loan/common/types"

const FEEDBACK_TABLENAME = "feedback"

// Feedback 描述与表 feedback 之间的映射
type Feedback struct {
	Id                    int64 `orm:"pk;"`
	AccountID             int64 `orm:"column(account_id)"`
	Mobile                string
	Content               string
	Tags                  int
	ApiVersion            string
	TaskVersion           string
	AppVersion            string
	AppVersionCode        int
	UIVersion             string `orm:"column(ui_version)"`
	CurrentOrderID        int64  `orm:"column(current_order_id)"`
	CurrentOrderStatus    types.LoanStatus
	CurrentOrderApplyTime int64
	ApplyOrderNum         int64
	ApplyOrderSuccNum     int64
	Status                int
	PhotoId1              int64 `orm:"column(photo_id1)"`
	PhotoId2              int64 `orm:"column(photo_id2)"`
	PhotoId3              int64 `orm:"column(photo_id3)"`
	PhotoId4              int64 `orm:"column(photo_id4)"`
	Ctime                 int64
	Utime                 int64
}

// TableName 返回当前模型对应的表名
func (r *Feedback) TableName() string {
	return FEEDBACK_TABLENAME
}

// Using 返回当前模型的数据库
func (r *Feedback) Using() string {
	return types.OrmDataBaseApi
}

func (r *Feedback) UsingSlave() string {
	return types.OrmDataBaseApiSlave
}
