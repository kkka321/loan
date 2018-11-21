package service

import (
	"fmt"

	"micro-loan/common/models"
	"micro-loan/common/tools"
	"micro-loan/common/types"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

func AddOneUploadResource(record map[string]interface{}) {
	obj := models.UploadResource{
		Id:         record["id"].(int64),
		OpUid:      record["op_uid"].(int64),
		ContentMd5: record["content_md5"].(string),
		HashName:   record["hash_name"].(string),
		Extension:  record["extension"].(string),
		UseMark:    record["use_mark"].(types.ResourceUseMark),
		Mime:       record["mime"].(string),
		Ctime:      tools.GetUnixMillis(),
	}

	o := orm.NewOrm()
	o.Using(obj.Using())
	o.Insert(&obj)
}

func OneResource(rid int64) (resource models.UploadResource, err error) {
	resource.Id = rid
	o := orm.NewOrm()
	o.Using(resource.Using())
	err = o.Read(&resource)

	if err != nil {
		logs.Warning("resource does not exsist. rid:", rid)
	}

	return
}

func BuildResourceUrl(resourceId int64) string {
	if resourceId <= 0 {
		return ""
	}

	//return fmt.Sprintf("%s/%s", AwsResouceDomain, obj.HashName)
	return fmt.Sprintf("/resource/img/%d", resourceId)
}
