package controllers

import (
	"micro-loan/common/cerror"
	"micro-loan/common/service"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
)

type AdvertisementController struct {
	beego.Controller
}

func (c *AdvertisementController) GetAdvertisement() {
	data, err := service.GetAdvertisement()
	if err != nil {
		logs.Error("Get advertisement to  fail err:", err)
		c.Data["json"] = cerror.BuildApiResponse(cerror.AdvertisementGetCodeFail, "")
		c.ServeJSON()
	}

	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	c.ServeJSON()
	return

}
