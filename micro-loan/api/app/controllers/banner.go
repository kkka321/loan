package controllers

import (
	"micro-loan/common/cerror"
	"micro-loan/common/service"
	"micro-loan/common/tools"
)

type BannerController struct {
	ApiBaseController
}

func (c *BannerController) Prepare() {
	// 调用上一级的 Prepare 方
	c.ApiBaseController.Prepare()
}

func (c *BannerController) GetBanners() {
	list, err := service.GetBanners()
	if err != nil {
		c.Data["json"] = cerror.BuildApiResponse(cerror.BannerGetCodeFail, "")
		c.ServeJSON()
		return
	}

	invite, err := service.GetInvite()
	if err != nil {
		c.Data["json"] = cerror.BuildApiResponse(cerror.BannerGetCodeFail, "")
		c.ServeJSON()
		return
	}

	data := map[string]interface{}{
		"server_time": tools.GetUnixMillis(),
		"banners":     list,
		"invite":      invite,
	}
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
}
