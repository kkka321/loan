package controllers

import (
	"fmt"
	"micro-loan/common/cerror"
	"micro-loan/common/pkg/feedback"
	"micro-loan/common/tools"
	"micro-loan/common/types"

	"github.com/astaxie/beego/logs"
)

type FeedbackController struct {
	ApiBaseController
}

func (c *FeedbackController) Prepare() {
	// 调用上一级的 Prepare 方
	c.ApiBaseController.Prepare()

	// 统一将 ip 加到 RequestJSON 中
	c.RequestJSON["ip"] = c.Ctx.Input.IP()
	c.RequestJSON["related_id"] = int64(0)
}

func (c *FeedbackController) Create() {
	// 查看是否可以复贷

	if !feedback.CheckCreateRequired(c.RequestJSON) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.LostRequiredParameters, "")
		c.ServeJSON()
		return
	}

	idList := make([]int64, 4)
	for i := 1; i <= 4; i++ {
		filename := fmt.Sprintf("fs%d", i)
		filesize := fmt.Sprintf("fs%d_size", i)

		_, ok := c.RequestJSON[filesize]
		if !ok {
			logs.Debug("[FeedbackController] UploadResource skip field fs:%s", filesize)
			break
		}

		resourceId, handHeldIdPhotoTmp, _, err := c.UploadResource(filename, types.Use2FeedbackPhoto)
		defer tools.Remove(handHeldIdPhotoTmp)
		if err != nil {
			logs.Error("[FeedbackController] UploadResource error name:%s, err:%v", filename, err)
			continue
		}

		logs.Debug("[FeedbackController] UploadResource success name:%s, resourceId:%d", filename, resourceId)

		idList[i-1] = resourceId
	}

	_, errCode := feedback.CreateByCustomer(c.AccountID, c.RequestJSON, idList)
	if errCode != cerror.CodeSuccess {
		c.Data["json"] = cerror.BuildApiResponse(errCode, "")
		c.ServeJSON()
		return
	}
	data := map[string]interface{}{
		"server_time": tools.GetUnixMillis(),
	}
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}
