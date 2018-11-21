package models

import (
	"fmt"
	"micro-loan/common/types"
	"strings"

	"github.com/astaxie/beego/orm"
)

const ACCOUNT_PROFILE_TABLENAME string = "account_profile"

// 一期只使用中/英文

// 工作类型
var jobTypeConf = map[int]string{
	1: "full time",
	2: "part time",
	3: "self-employed",
	4: "no-work",
}

// 月收入
var monthlyIncomeConf = map[int]string{
	1: "di bawah 1jt",
	2: "1-3jt,",
	3: "3-5jt",
	4: "5-10jt",
	5: "10-20jt",
	6: "20jt ke atas",
}

// 工作年限
var serviceYearsConf = map[int]string{
	1: "Within 3 months",
	2: "Within 6 months",
	3: "Within 1 year",
	4: "Within 2 years",
	5: "2 years or more",
}

// 联系人关系
var relationshipConf = map[int]string{
	1: "parents",   // 父母
	2: "spouse",    // 配偶
	3: "child",     // 子女
	4: "friend",    // 朋友
	5: "colleague", // 同事
	6: "other",     // 其它
}

// 教育状况
var educationConf = map[int]string{
	1: "Undergraduate and above",
	2: "High school",
	3: "Secondary school",
	4: "Junior high school and below",
}

// 婚姻状况
var maritalStatusConf = map[int]string{
	1: "unmarried",
	2: "married",
	3: "Divorced",
}

// 子女数
var childrenNumberConf = map[int]string{
	-1: "Undefined",
	0:  "None",
	1:  "1",
	2:  "2",
	3:  "3",
	4:  "more than 3",
}

type AccountProfile struct {
	AccountId                 int64  `orm:"pk;column(account_id)"`
	IdPhoto                   int64  `orm:"column(id_photo)"`
	HandHeldIdPhoto           int64  `orm:"column(hand_held_id_photo)"`
	JobType                   int    `orm:"column(job_type)"`
	MonthlyIncome             int    `orm:"column(monthly_income)"`
	CompanyName               string `orm:"column(company_name)"`
	CompanyCity               string `orm:"column(company_city)"`
	CompanyAddress            string `orm:"column(company_address)"`
	ServiceYears              int    `orm:"column(service_years)"`
	Contact1                  string
	Contact1Name              string `orm:"column(contact1_name)"`
	Relationship1             int
	Contact2                  string
	Contact2Name              string `orm:"column(contact2_name)"`
	Relationship2             int
	Education                 int
	MaritalStatus             int    `orm:"column(marital_status)"`
	ChildrenNumber            int    `orm:"column(children_number)"`
	ResidentCity              string `orm:"column(resident_city)"`
	ResidentAddress           string `orm:"column(resident_address)"`
	BankName                  string `orm:"column(bank_name)"`
	BankNo                    string `orm:"column(bank_no)"`
	RepayBankCode             string `orm:"column(repay_bank_code)"`
	IdPhotoQuality            string `orm:"column(id_photo_quality)"`
	IdPhotoQualityThreshold   string `orm:"column(id_photo_quality_threshold)"`
	HandPhotoQuality          string `orm:"column(hand_photo_quality)"`
	HandPhotoQualityThreshold string `orm:"column(hand_photo_quality_threshold)"`
	FaceComparison            string `orm:"column(face_comparison)"`
	FaceComparisonIhID        string `orm:"column(face_comparison_ih_id)"`
	IdHoldingPhotoCheck       string `orm:"column(id_holding_photo_check)"`
	CompanyTelephone          string
	SalaryDay                 string
	Ctime                     int64
	Utime                     int64
}

func (r *AccountProfile) TableName() string {
	return ACCOUNT_PROFILE_TABLENAME
}

func (r *AccountProfile) Using() string {
	return types.OrmDataBaseApi
}
func (r *AccountProfile) UsingSlave() string {
	return types.OrmDataBaseApiSlave
}

//! 这个方法有点副作用,请调用者注意
func (r *AccountProfile) Update(cols ...string) (id int64, err error) {
	o := orm.NewOrm()
	o.Using(r.Using())

	id, err = o.Update(r, cols...)

	return
}

func (r *AccountProfile) Delete(cols ...string) (id int64, err error) {
	o := orm.NewOrm()
	o.Using(r.Using())

	id, err = o.Delete(r, cols...)

	return
}

// 更改银行账号
func (r *AccountProfile) ChangeBankNo(bankNo string) (err error) {
	o := orm.NewOrm()
	o.Using(r.Using())

	r.BankNo = bankNo
	_, err = o.Update(r, "bank_no")

	return
}

// 保存人脸识别结果
func (r *AccountProfile) SaveIdPhotoDetect(quality, qualityThreshold float64) (err error) {
	o := orm.NewOrm()
	o.Using(r.Using())

	r.IdPhotoQuality = fmt.Sprintf("%.14f", quality)
	r.IdPhotoQualityThreshold = fmt.Sprintf("%f", qualityThreshold)
	_, err = o.Update(r, "id_photo_quality", "id_photo_quality_threshold")

	return
}

// 保存手持照片人脸识别结果
func (r *AccountProfile) SaveHandPhotoDetect(quality, qualityThreshold float64) (err error) {
	o := orm.NewOrm()
	o.Using(r.Using())

	r.HandPhotoQuality = fmt.Sprintf("%.14f", quality)
	r.HandPhotoQualityThreshold = fmt.Sprintf("%f", qualityThreshold)
	_, err = o.Update(r, "hand_photo_quality", "hand_photo_quality_threshold")

	return
}

// 保存人脸对比结果
func (r *AccountProfile) SaveFaceComparison(faceCmp float64) (err error) {
	o := orm.NewOrm()
	o.Using(r.Using())

	r.FaceComparison = fmt.Sprintf("%.14f", faceCmp)
	_, err = o.Update(r, "face_comparison")

	return
}

// 保存手持识别结果
func (r *AccountProfile) SaveHoldCheck(holdCheck float64) (err error) {
	o := orm.NewOrm()
	o.Using(r.Using())

	r.IdHoldingPhotoCheck = fmt.Sprintf("%.14f", holdCheck)
	_, err = o.Update(r, "id_holding_photo_check")

	return
}

// 保存手持与身份证比对结果
func (r *AccountProfile) SaveHoldAndIDComparison(holdCheck float64) (err error) {
	o := orm.NewOrm()
	o.Using(r.Using())

	r.FaceComparisonIhID = fmt.Sprintf("%.14f", holdCheck)
	_, err = o.Update(r, "face_comparison_ih_id")

	return
}

func GetJobTypeConf() map[int]string {
	return jobTypeConf
}

func GetMonthlyIncomeConf() map[int]string {
	return monthlyIncomeConf
}

func GetServiceYearsConf() map[int]string {
	return serviceYearsConf
}

func GetRelationshipConf() map[int]string {
	return relationshipConf
}

func GetEducationConf() map[int]string {
	return educationConf
}

func GetMaritalStatusConf() map[int]string {
	return maritalStatusConf
}

func GetChildrenNumberConf() map[int]string {
	return childrenNumberConf
}

// OneAccountProfileByAccountID 使用 accountID 查询用户信息
func OneAccountProfileByAccountID(id int64) (AccountProfile, error) {
	var obj = AccountProfile{
		AccountId: id,
	}
	o := orm.NewOrm()
	o.Using(obj.Using())
	err := o.Read(&obj)

	return obj, err
}

// 取居住地省份
func (r *AccountProfile) ResidentProvince() (province string, err error) {
	if len(r.ResidentCity) <= 0 {
		err = fmt.Errorf("[*AccountProfile->ResidentProvince] has no `resident_city` data")
		return
	}

	box := strings.Split(r.ResidentCity, ",")
	if len(box) < 2 {
		err = fmt.Errorf("[*AccountProfile->ResidentProvince] `resident_city` is invalid")
		return
	}

	province = box[0]

	return
}

// 取公司地省份
func (r *AccountProfile) CompanyProvince() (province string, err error) {
	if len(r.CompanyCity) <= 0 {
		err = fmt.Errorf("[*AccountProfile->CompanyProvince] has no `company_city` data")
		return
	}

	box := strings.Split(r.CompanyCity, ",")
	if len(box) < 2 {
		err = fmt.Errorf("[*AccountProfile->CompanyProvince] `company_city` is invalid")
		return
	}

	province = box[0]

	return
}
