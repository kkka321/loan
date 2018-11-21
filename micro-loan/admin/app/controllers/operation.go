package controllers

import (
	"fmt"
	"micro-loan/common/models"
	"micro-loan/common/service"
	"micro-loan/common/tools"
	"micro-loan/common/types"

	"github.com/astaxie/beego/logs"
)

type OperationController struct {
	BaseController
}

func (c *OperationController) Prepare() {
	// 调用上一级的 Prepare 方法
	c.BaseController.Prepare()

	c.Data["Controller"] = "operation"
}

//
func (c *OperationController) ListAdvertisement() {
	c.Data["Action"] = "list"
	c.TplName = "operation/advertisement_list.html"

	list, err := service.GetMultiAdvertisements()
	if err != nil {
		c.commonError("", "advertisement_list", "advertisement  failed")
		return
	}
	c.Data["List"] = list

	c.Layout = "layout.html"
	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "operation/advertisement_list.js.html"
	return
}

func (c *OperationController) AddAdvertisement() {
	action := ""
	url := "list_advertisement"
	linkUrl := c.GetString("link_url")
	if len(linkUrl) <= 0 {
		linkUrl = ""
	}

	sourcePage, _ := c.GetInt64("source_page")
	if sourcePage <= 0 {
		sourcePage = 0
	}

	sTimeRange := c.GetString("stime_range")
	var sTime int64
	if len(sTimeRange) > 0 {
		timeStart := tools.GetDateParseBackends(sTimeRange) * 1000
		sTime = timeStart
	} else {
		c.commonError(action, url, "input start time failed")
		return
	}

	eTimeRange := c.GetString("etime_range")
	var eTime int64
	if len(eTimeRange) > 0 {
		timeEnd := tools.GetDateParseBackends(eTimeRange) * 1000
		eTime = timeEnd
	} else {
		c.commonError(action, url, "input end time  failed")
		return
	}

	times, err := service.GetMultiAdvertisements()
	if err != nil && err.Error() != types.EmptyOrmStr {
		c.commonError(action, url, "add advertisment failed")
		return
	}
	for _, v := range times {
		if (sTime >= v.StartTm && sTime <= v.EndTm) || (eTime >= v.StartTm && eTime <= v.EndTm) {
			c.commonError(action, url, "please check the time range")
			return
		}
	}
	c.Data["stimeRange"] = sTimeRange
	c.Data["etimeRange"] = eTimeRange

	fileNum, _ := c.GetInt("file_num")
	if fileNum <= 0 {
		c.commonError(action, url, " file is empty")
		return
	}

	resId, err := addPic(c, types.Use2Advertisement)
	if err != nil {
		logs.Error("[AddAdvertisement] failed, err is", err)
		c.commonError(action, url, "advertisementPic failed")
		return
	}
	rest := models.Advertisement{
		ResourceId: resId,
		LinkUrl:    linkUrl,
		SourcePage: sourcePage,
		StartTm:    sTime,
		EndTm:      eTime,
		IsShow:     1,
		Ctime:      tools.GetUnixMillis(),
		Utime:      tools.GetUnixMillis(),
	}
	_, errs := rest.Insert()
	if errs != nil {
		logs.Error("[AddAdvertisement] insert failed, err is", err)
		c.commonError(action, url, "advertisement insert failed")
		return
	}
	c.Redirect("/operation/list_advertisement", 302)
}

func (c *OperationController) UpdateAdvertisement() {
	action := ""
	url := "list_advertisement"
	id, _ := c.GetInt64("Id")

	isShow, _ := c.GetInt64("is_show")

	update := models.Advertisement{
		Id:     id,
		IsShow: isShow,
		Utime:  tools.GetUnixMillis(),
	}
	_, errs := update.Updates("id", "is_show", "utime")
	if errs != nil {
		logs.Error("[AddAdvertisement] update failed, err is", errs)
		c.commonError(action, url, "advertisement update failed")
		return
	}

	c.Redirect("/operation/list_advertisement", 302)
}

func (c *OperationController) DelAdvertisement() {

	mapData := make(map[string]interface{})
	mapData["data"] = false

	id, _ := c.GetInt64("Id")

	del := models.Advertisement{
		Id: id,
	}
	_, errs := del.Dels("id")
	if errs != nil {
		mapData["data"] = false
	} else {
		mapData["data"] = true
	}
	c.Data["json"] = &mapData
	c.ServeJSON()

}

func (c *OperationController) ListBanner() {
	c.Data["Action"] = "list"
	c.TplName = "operation/banner_list.html"

	list, err := service.GetMultiBanners()
	if err != nil {
		c.commonError("", "banner_list", "list banner  failed")
		return
	}
	c.Data["List"] = list

	bannerType, err := c.GetInt("banner_type")
	if err == nil && bannerType >= 0 {
		c.Data["bannerType"] = bannerType
	} else {
		c.Data["bannerType"] = -1
	}
	c.Data["bannerTypeList"] = types.BannerTypeMap

	c.Layout = "layout.html"
	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "operation/banner_list.js.html"
	return
}

func (c *OperationController) AddBanner() {

	action := ""
	url := "list_banner"

	rId, _ := c.GetInt64("r_id") //resource_id

	linkUrl := c.GetString("link_url")
	if len(linkUrl) <= 0 {
		linkUrl = ""
	}

	sourcePage, _ := c.GetInt64("source_page")
	if sourcePage <= 0 {
		sourcePage = 0
	}

	postion, _ := c.GetInt64("postion")
	if postion <= 0 {
		postion = 0
	}

	var resId int64
	fileNum, _ := c.GetInt("file_num")

	if rId <= 0 && fileNum <= 0 {
		logs.Error("rId and fileNum is 0")
		c.commonError(action, url, "The file is not empty")
		return
	}

	if fileNum <= 0 {
		resId = rId
	} else {
		resIds, err := addPic(c, types.Use2Banner)
		if err != nil {
			logs.Error("[AddBanner] failed, err is", err, resId)
			c.commonError(action, url, "bannerPic failed")
			return
		}
		resId = resIds
	}

	bannerType, err := c.GetInt("banner_type")
	if err == nil && bannerType >= 0 {
		c.Data["bannerType"] = bannerType
	} else {
		c.Data["bannerType"] = -1
		c.commonError(action, url, "The bannerType doesn't select")
		return
	}

	var startTimeStamp int64
	startTime := c.GetString("start_time")
	if len(startTime) <= 0 {
		startTime = ""
		c.commonError(action, url, "The start time is empty")
		return
	} else {
		startTimeStamp, _ = tools.GetTimeParseWithFormat(startTime, "2006-01-02 15:04:05")
		startTimeStamp *= 1000
	}

	var endTimeStamp int64
	endTime := c.GetString("end_time")
	if len(endTime) <= 0 {
		endTime = ""
		c.commonError(action, url, "The end time is empty")
		return
	} else {
		endTimeStamp, _ = tools.GetTimeParseWithFormat(endTime, "2006-01-02 15:04:05")
		endTimeStamp *= 1000
	}

	content := c.GetString("content")
	if len(content) <= 0 {
		content = ""
	}

	fontColor := c.GetString("font_color")
	if len(fontColor) <= 0 {
		fontColor = ""
	}

	fontLintUrl := c.GetString("font_link_url")
	if len(fontLintUrl) <= 0 {
		fontLintUrl = ""
	}

	rest := models.Banner{
		ResourceId:  resId,
		LinkUrl:     linkUrl,
		SourcePage:  sourcePage,
		Postion:     postion,
		BannerType:  bannerType,
		StartTime:   startTimeStamp,
		EndTime:     endTimeStamp,
		Content:     content,
		FontColor:   fontColor,
		FontLinkUrl: fontLintUrl,
		Ctime:       tools.GetUnixMillis(),
		Utime:       tools.GetUnixMillis(),
	}

	if rId <= 0 {
		_, errs := rest.Insert()
		if errs != nil {
			logs.Error("[AddBanner] insert failed, err is", errs)
			c.commonError(action, url, "banner insert failed")
			return
		}
	} else {
		ids, _ := c.GetInt64("ids")
		update := models.Banner{
			Id:          ids,
			ResourceId:  resId,
			LinkUrl:     linkUrl,
			SourcePage:  sourcePage,
			Postion:     postion,
			BannerType:  bannerType,
			StartTime:   startTimeStamp,
			EndTime:     endTimeStamp,
			Content:     content,
			FontColor:   fontColor,
			FontLinkUrl: fontLintUrl,
			Utime:       tools.GetUnixMillis(),
		}

		_, errs := update.Updates("id", "resource_id", "link_url", "source_page", "postion", "type",
			"start_time", "end_time", "content", "font_color", "font_link_url", "utime")
		if errs != nil {
			logs.Error("[UpdateBanner] update failed, err is", errs)
			c.commonError(action, url, "banner update failed")
			return
		}
	}

	c.Redirect("/operation/list_banner", 302)
}

func (c *OperationController) DelBanner() {

	mapData := make(map[string]interface{})
	mapData["data"] = false

	id, _ := c.GetInt64("Id")
	del := models.Banner{
		Id: id,
	}
	_, errs := del.Dels("id")
	if errs != nil {
		mapData["data"] = false
	} else {
		mapData["data"] = true
	}

	c.Data["json"] = &mapData
	c.ServeJSON()
}

func (c *OperationController) UpdateBannerPostion() {
	action := ""
	url := "list_banner"

	mapData := make(map[string]interface{})
	mapData["data"] = false

	id := c.GetStrings("[]Id")
	postion := c.GetStrings("[]Postion")

	for i := 0; i < len(id); i++ {
		ids, _ := tools.Str2Int64(id[i])
		postions, _ := tools.Str2Int64(postion[i])

		update := models.Banner{
			Id:      ids,
			Postion: postions,
			Utime:   tools.GetUnixMillis(),
		}

		_, errs := update.Updates("id", "postion", "utime")
		if errs != nil {
			logs.Error("[UpdateBanner] update failed, err is", errs)
			c.commonError(action, url, "banner update postion failed")
			return
		}
	}

	c.Layout = "layout.html"
	c.TplName = "operation/banner_list.html"
	c.Redirect("/operation/list_banner", 302)

}

func addPic(c *OperationController, useMark types.ResourceUseMark) (int64, error) {
	fileName := fmt.Sprintf("file%d", 0)
	resId, idPicTmp, code, err := c.UploadResource(fileName, useMark)
	logs.Debug("[refundToBankCard] idPhoto:%d, idPhotoTmp:%s, code:%d, err:%v fileName:%s", resId, idPicTmp, code, err, fileName)
	defer tools.Remove(idPicTmp)
	if err != nil {
		err = fmt.Errorf("[refundToBankCard] 上传凭证失败  update pic err:%v", err)
		logs.Error(err)
		return 0, err
	}

	return resId, nil
}

func (c *OperationController) ListAdPosition() {
	c.Data["Action"] = "list"
	c.TplName = "operation/ad_position_list.html"

	list, err := service.GetAdPositionList()
	if err != nil {
		c.commonError("", "ad_position_list", "list ad_position  failed")
		return
	}
	c.Data["List"] = list

	position, err := c.GetInt("position")
	if err == nil && position >= 0 {
		c.Data["position"] = position
	} else {
		c.Data["position"] = -1
	}
	c.Data["adPositionList"] = types.AdPositionMap

	c.Layout = "layout.html"
	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "operation/ad_position_list.js.html"
	return
}

func (c *OperationController) AddAdPosition() {

	action := ""
	url := "list_ad_position"

	rId, _ := c.GetInt64("r_id") //resource_id

	linkUrl := c.GetString("link_url")
	if len(linkUrl) <= 0 {
		linkUrl = ""
	}

	companyId, _ := c.GetInt("company_id")
	if companyId <= 0 {
		companyId = 0
	}

	position, _ := c.GetInt("position")
	if position <= 0 {
		position = 0
	}

	var resId int64
	fileNum, _ := c.GetInt("file_num")

	if rId <= 0 && fileNum <= 0 {
		logs.Error("rId and fileNum is 0")
		c.commonError(action, url, "The file is not empty")
		return
	}

	if fileNum <= 0 {
		resId = rId
	} else {
		resIds, err := addPic(c, types.Use2AdPosition)
		if err != nil {
			logs.Error("[AddAdPosition] failed, err is", err, resId)
			c.commonError(action, url, "adPosionPic failed")
			return
		}
		resId = resIds
	}

	t := tools.GetUnixMillis()
	rest := models.AdPosition{
		ResourceId: resId,
		LinkUrl:    linkUrl,
		CompanyId:  companyId,
		Position:   position,
		Ctime:      t,
		Utime:      t,
	}

	if rId <= 0 {
		_, errs := rest.Insert()
		if errs != nil {
			logs.Error("[AddAdPosition] insert failed, err is", errs)
			c.commonError(action, url, "ad position insert failed")
			return
		}
	} else {
		ids, _ := c.GetInt64("ids")
		update := models.AdPosition{
			Id:         ids,
			ResourceId: resId,
			LinkUrl:    linkUrl,
			CompanyId:  companyId,
			Position:   position,
			Utime:      tools.GetUnixMillis(),
		}

		_, errs := update.Updates("id", "resource_id", "link_url", "company_id", "position", "utime")
		if errs != nil {
			logs.Error("[UpdateAdPosition] update failed, err is", errs)
			c.commonError(action, url, "ad position update failed")
			return
		}
	}

	c.Redirect("/operation/list_ad_position", 302)
}

func (c *OperationController) DelAdPosition() {

	mapData := make(map[string]interface{})
	mapData["data"] = false

	id, _ := c.GetInt64("Id")
	del := models.AdPosition{
		Id: id,
	}
	_, errs := del.Dels("id")
	if errs != nil {
		mapData["data"] = false
	} else {
		mapData["data"] = true
	}

	c.Data["json"] = &mapData
	c.ServeJSON()
}
