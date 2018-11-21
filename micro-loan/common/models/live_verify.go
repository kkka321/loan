package models

import (
	"micro-loan/common/types"
)

const LIVE_VERIFY_TABLENAME string = "live_verify"

type LiveVerify struct {
	Id             int64   `orm:"pk"`
	AccountId      int64   `orm:"column(account_id)"`
	OrderID        int64   `orm:"column(order_id)"`
	ImageBest      int64   `orm:"column(image_best)"`
	ImageEnv       int64   `orm:"column(image_env)"`
	ImageRef1      int64   `orm:"column(image_ref1)"`
	ConfidenceRef1 float64 `orm:"column(confidence_ref1)"`
	ImageRef2      int64   `orm:"column(image_ref2)"`
	ConfidenceRef2 float64 `orm:"column(confidence_ref2)"`
	ImageRef3      int64   `orm:"column(image_ref3)"`
	ConfidenceRef3 float64 `orm:"column(confidence_ref3)"`
	Ctime          int64
	// "confidence": Float, in [0，100]. Higher confidence indicates higher possibility that two faces belong to same person. This value should be compared with one of the threshold below. If it is larger than the threshold, you can believe the two faces are of on person.
	//Confidence string // 会有多个值,以`,`隔开.形如: result_ref1:93.025,result_ref2:88.982,result_ref3:93.085
}

func (r *LiveVerify) TableName() string {
	return LIVE_VERIFY_TABLENAME
}

func (r *LiveVerify) Using() string {
	return types.OrmDataBaseApi
}
func (r *LiveVerify) UsingSlave() string {
	return types.OrmDataBaseApiSlave
}

func (r *LiveVerify) VerifyConfidence() float64 {
	return (r.ConfidenceRef1 + r.ConfidenceRef2 + r.ConfidenceRef3) / 3
}
