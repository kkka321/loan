package controllers

import (
	"fmt"
	"micro-loan/common/models"
	"micro-loan/common/pkg/feedback"
	"micro-loan/common/service"
	"micro-loan/common/tools"
	"micro-loan/common/types"

	"github.com/astaxie/beego/logs"
)

type ActivityController struct {
	BaseController
}

func (c *ActivityController) Prepare() {
	// 调用上一级的 Prepare 方法
	c.BaseController.Prepare()

	c.Data["Controller"] = "operation"
}

func (c *ActivityController) ListFloating() {
	c.Data["Action"] = "list"
	c.TplName = "activity/floating_list.html"

	list, err := service.GetFloatingBg()
	if err != nil {
		c.commonError("", "floating_list", "list banner  failed")
		return
	}
	c.Data["List"] = list

	c.Data["tagFloatingMap"] = feedback.TagsFloatingMap()
	c.Layout = "layout.html"
	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "activity/floating_list.js.html"
	return
}

func (c *ActivityController) AddFloating() {

	action := ""
	url := "list_floating"

	rId, _ := c.GetInt64("r_id") //resource_id

	linkUrl := c.GetString("link_url")
	if len(linkUrl) <= 0 {
		linkUrl = ""
	}

	sourcePage, _ := c.GetInt64("source_page")
	if sourcePage <= 0 {
		sourcePage = 0
	}

	floatingTags, _ := c.GetInt64("floating_tags")
	if floatingTags <= 0 {
		logs.Error("[AddFloating] floatingTags  is 0")
		c.commonError(action, url, "floating is not empty")
		return
	}
	c.Data["floating_tags"] = floatingTags

	if sourcePage == 0 && len(linkUrl) == 0 {
		c.commonError(action, url, "url can't be empty failed")
		return
	}
	if rId <= 0 {
		res, err := models.GetOneByEtypeAndPostionPopWindow(1,floatingTags)
		if err != nil && err.Error() != types.EmptyOrmStr {
			logs.Error("[AddFloating] models GetOneByEtypePopWindow err :%v ", err)
			c.commonError(action, url, "floating get failed")
			return
		}
		if len(res) > 0 {
			logs.Error("[AddFloating] models GetOneByEtypePopWindow err  already advertising Spaces")
			c.commonError(action, url, "There are already floating Spaces")
			return
		}
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
		resIds, err := addPics(c, types.Use2Float)
		if err != nil {
			logs.Error("[AddFloating] failed, err is", err, resId)
			c.commonError(action, url, "floating add Pic failed")
			return
		}
		resId = resIds
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

	if eTime <= sTime {
		c.commonError(action, url, "end time  be greater than or equal start time")
		return
	}

	/*nowDate := tools.GetUnixMillis()
	if sTime <= nowDate {
		c.commonError(action, url, "start time Must be greater than now time ")
		return
	}*/
	// times, err := service.GetFloatingBg()
	// if err != nil && err.Error() != types.EmptyOrmStr {
	// 	c.commonError(action, url, "add floating failed")
	// 	return
	// }
	// for _, v := range times {
	// 	if (sTime >= v.StartTm && sTime <= v.EndTm) || (eTime >= v.StartTm && eTime <= v.EndTm) {
	// 		c.commonError(action, url, "please check the time range")
	// 		return
	// 	}
	// }
	c.Data["stimeRange"] = sTimeRange
	c.Data["etimeRange"] = eTimeRange

	rest := models.Activity{
		Etype:      1,
		FPostion:   floatingTags,
		ResourceId: resId,
		LinkUrl:    linkUrl,
		SourcePage: sourcePage,
		StartTm:    sTime,
		EndTm:      eTime,
		IsShow:     1,
		Ctime:      tools.GetUnixMillis(),
		Utime:      tools.GetUnixMillis(),
	}

	if rId <= 0 {
		_, errs := rest.Insert()
		if errs != nil {
			logs.Error("[AddFloating] insert failed, err is", errs)
			c.commonError(action, url, "floating insert failed")
			return
		}
	} else {
		ids, _ := c.GetInt64("ids")
		update := models.Activity{
			Id:         ids,
			Etype:      1,
			FPostion:   floatingTags,
			ResourceId: resId,
			LinkUrl:    linkUrl,
			SourcePage: sourcePage,
			StartTm:    sTime,
			EndTm:      eTime,
			Utime:      tools.GetUnixMillis(),
		}

		_, errs := update.Updates("id", "etype", "f_postion", "resource_id", "link_url", "source_page", "start_tm", "end_tm", "utime")
		if errs != nil {
			logs.Error("[AddFloating] update failed, err is", errs)
			c.commonError(action, url, "floating update failed")
			return
		}
	}
	c.Data["tagFloatingMap"] = feedback.TagsFloatingMap()
	c.Redirect("/activity/list_floating", 302)
}

func (c *ActivityController) DelFloating() {

	mapData := make(map[string]interface{})
	mapData["data"] = false

	id, _ := c.GetInt64("Id")
	del := models.Activity{
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

func (c *ActivityController) ListPopUpWindow() {
	c.Data["Action"] = "list"
	c.TplName = "activity/pop_list.html"

	list, err := service.GetPopBg()
	if err != nil {
		c.commonError("", "pop_list", "list pop up window  failed")
		return
	}
	c.Data["List"] = list
	c.Data["tagFloatingMap"] = feedback.TagsFloatingMap()
	c.Layout = "layout.html"
	c.LayoutSections = make(map[string]string)
	c.LayoutSections["Scripts"] = "activity/pop_list.js.html"
	return
}

func (c *ActivityController) AddPopUpWindow() {

	action := ""
	url := "list_popupwindow"

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

	if sourcePage == 0 && len(linkUrl) == 0 {
		c.commonError(action, url, "url can't be empty failed")
		return
	}

	if rId <= 0 {
		res, err := models.GetOneByEtypePopWindow(0)
		if err != nil && err.Error() != types.EmptyOrmStr {
			logs.Error("[AddPopUpWindow] models GetOneByEtypeAndPostionPopWindow err :%v ", err)
			c.commonError(action, url, "floating get failed")
			return
		}
		if len(res) > 0 {
			logs.Error("[AddPopUpWindow] models GetOneByEtypeAndPostionPopWindow err  already advertising Spaces")
			c.commonError(action, url, "There are already advertising Spaces")
			return
		}
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
		resIds, err := addPics(c, types.Use2Pop)
		if err != nil {
			logs.Error("[AddPopUpWindow] failed, err is", err, resId)
			c.commonError(action, url, "pop up window add Pic failed")
			return
		}
		resId = resIds
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
	if eTime <= sTime {
		c.commonError(action, url, " end time  be greater than or equal start time")
		return
	}
	/*nowDate := tools.GetUnixMillis()
	if sTime <= nowDate {
		c.commonError(action, url, "start time Must be greater than now time ")
		return
	}*/

	// times, err := service.GetPopBg()
	// if err != nil && err.Error() != types.EmptyOrmStr {
	// 	c.commonError(action, url, "add pop up window failed")
	// 	return
	// }
	// for _, v := range times {
	// 	if (sTime >= v.StartTm && sTime <= v.EndTm) || (eTime >= v.StartTm && eTime <= v.EndTm) {
	// 		c.commonError(action, url, "please check the time range")
	// 		return
	// 	}
	// }
	c.Data["stimeRange"] = sTimeRange
	c.Data["etimeRange"] = eTimeRange

	rest := models.Activity{
		Etype:      0,
		ResourceId: resId,
		LinkUrl:    linkUrl,
		SourcePage: sourcePage,
		StartTm:    sTime,
		EndTm:      eTime,
		IsShow:     1,
		Ctime:      tools.GetUnixMillis(),
		Utime:      tools.GetUnixMillis(),
	}

	if rId <= 0 {
		_, errs := rest.Insert()
		if errs != nil {
			logs.Error("[AddPopUpWindow] insert failed, err is", errs)
			c.commonError(action, url, "floating insert failed")
			return
		}
	} else {
		ids, _ := c.GetInt64("ids")
		update := models.Activity{
			Id:         ids,
			Etype:      0,
			ResourceId: resId,
			LinkUrl:    linkUrl,
			SourcePage: sourcePage,
			StartTm:    sTime,
			EndTm:      eTime,
			Utime:      tools.GetUnixMillis(),
		}

		_, errs := update.Updates("id", "etype", "f_postion", "resource_id", "link_url", "source_page", "start_tm", "end_tm", "utime")
		if errs != nil {
			logs.Error("[AddPopUpWindow] update failed, err is", errs)
			c.commonError(action, url, "pop uo window update failed")
			return
		}
	}

	c.Redirect("/activity/list_popupwindow", 302)
}

// func (c *ActivityController) UpdatePopUpWindow() {
// 	action := ""
// 	url := "list_popupwindow"

// 	mapData := make(map[string]interface{})
// 	mapData["data"] = false

// 	id := c.GetStrings("[]Id")
// 	postion := c.GetStrings("[]Postion")
// 	for i := 0; i < len(id); i++ {
// 		ids, _ := tools.Str2Int64(id[i])
// 		postions, _ := tools.Str2Int64(postion[i])

// 		update := models.Activity{
// 			Id:    ids,
// 			Utime: tools.GetUnixMillis(),
// 		}

// 		_, errs := update.Updates("id", "postion", "utime")
// 		if errs != nil {
// 			logs.Error("[UpdatePopUpWindow] update failed, err is", errs)
// 			c.commonError(action, url, "activity pop up window update failed")
// 			return
// 		}
// 	}

// 	c.Layout = "layout.html"
// 	c.TplName = "/activity/list_popupwindow.html"
// 	c.Redirect("/activity/list_popupwindow", 302)

// }

func (c *ActivityController) DelPopUpWindow() {

	mapData := make(map[string]interface{})
	mapData["data"] = false

	id, _ := c.GetInt64("Id")

	del := models.Activity{
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

func addPics(c *ActivityController, useMark types.ResourceUseMark) (int64, error) {
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
