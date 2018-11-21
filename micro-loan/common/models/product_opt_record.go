package models

// `product_opt_record`
import (
	"github.com/astaxie/beego/orm"

	"micro-loan/common/types"
)

const PRODUCT_OPT_RECORD_TABLENAME string = "product_opt_record"

type ProductOptRecord struct {
	Id          int64 `orm:"pk;"`
	ProductId   int64
	ProductName string
	Nickname    string
	OpType      int
	OpUid       int64
	Original    string
	Edited      string
	Ctime       int64
}

// TableName 当前模型对应的表名
func (r *ProductOptRecord) TableName() string {
	return PRODUCT_OPT_RECORD_TABLENAME
}

// Using 当前模型的数据库
func (r *ProductOptRecord) Using() string {
	return types.OrmDataBaseAdmin
}
func (r *ProductOptRecord) UsingSlave() string {
	return types.OrmDataBaseAdminSlave
}

func (r *ProductOptRecord) AddRecord() (int64, error) {
	o := orm.NewOrm()
	o.Using(r.Using())
	id, err := o.Insert(r)
	return id, err
}

func GetProductOptRecordByPkId(id int64) (pro ProductOptRecord, err error) {
	o := orm.NewOrm()
	o.Using(pro.Using())

	err = o.QueryTable(pro.TableName()).
		Filter("Id", id).
		One(&pro)
	return
}

// 此表时操作流水，不允许update

// func (r *ProductOptRecord) UpdateProduct(cols ...string) (id int64, err error) {
// 	o := orm.NewOrm()
// 	o.Using(r.Using())

// 	id, err = o.Update(r, cols...)

// 	return
// }

// func GetProduct(id int64) (Product, error) {

// 	var product Product

// 	o := orm.NewOrm()
// 	o.Using(product.Using())
// 	err := o.QueryTable(product.TableName()).Filter("Id", id).One(&product)

// 	return product, err
// }
