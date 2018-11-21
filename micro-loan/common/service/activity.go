package service

import (
	"fmt"
	"micro-loan/common/models"
	"micro-loan/common/tools"
	"micro-loan/common/types"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
)

type ActivityResponse struct {
	PicUrl     string
	LinkUrl    string
	SourcePage int64
}

func GetPopoversor() (map[string]interface{}, error) {
	res, err := models.GetOneByEtypeAndPostionPopWindow(0, 0)
	if err != nil && err.Error() != types.EmptyOrmStr {
		logs.Error("[GetPopoversor] models GetOneByEtypeAndPostionPopWindow err :%v ", err)
		return nil, err
	}
	is_pop := false
	mps := make([]map[string]interface{}, 0)
	for _, v := range res {
		picUrl := ""
		pic, err := OneResource(v.ResourceId)
		if err != nil {
			logs.Error("[GetAdvertisement] models OneResource err :%v ", err)
			return nil, err
		}

		picUrl = fmt.Sprintf("%s/%s", beego.AppConfig.String("ad_cdn_url"), pic.HashName)

		if v.ResourceId > 0 {
			is_pop = true
		}

		mp := map[string]interface{}{
			"ad_url":      picUrl,
			"link_url":    v.LinkUrl,
			"source_page": v.SourcePage,
			"server_time": tools.GetUnixMillis(),
		}
		mps = append(mps, mp)
		is_pop = true
	}
	resMp := map[string]interface{}{
		"is_pop": is_pop,
		"list":   mps,
	}
	return resMp, nil
}

func GetFloating(fPostion string) (map[string]interface{}, error) {
	postion, _ := tools.Str2Int64(fPostion)

	data, err := models.GetOneByEtypeAndPostionFloating(1, postion)
	if err != nil && err.Error() != types.EmptyOrmStr {
		logs.Error("[GetFloating] models GetOneByEtypeAndPostionFloating err :%v ", err)
		return nil, err
	}
	picUrl := ""
	if data != (models.Activity{}) {
		pic, err := OneResource(data.ResourceId)
		if err != nil {
			logs.Error("[GetFloating] models OneResource err :%v ", err)
			return nil, err
		}

		picUrl = fmt.Sprintf("%s/%s", beego.AppConfig.String("ad_cdn_url"), pic.HashName)

	} else {
		data = models.Activity{}
	}

	is_ft := false
	if data.ResourceId > 0 {
		is_ft = true
	}

	mp := map[string]interface{}{
		"is_ft":       is_ft,
		"ad_url":      picUrl,
		"link_url":    data.LinkUrl,
		"source_page": data.SourcePage,
		"server_time": tools.GetUnixMillis(),
	}
	return mp, nil
}

type ActivityList struct {
	Id         int64
	Etype      int64
	FPostion   int64
	ResourceId int64
	LinkUrl    string
	SourcePage int64
	StartTm    int64
	EndTm      int64
	IsShow     int64
	Ctime      int64
	Utime      int64
	PicUrl     string
}
func GetFloatingBg() ([]ActivityList, error) {

	data, err := models.GetAllByEtype(1)
	if err != nil && err.Error() != types.EmptyOrmStr {
		logs.Error("[GetFloating] models GetOneByEtypeAndPostion err :%v ", err)
		return nil, err
	}

	res:= make([]ActivityList,0)
	for _, v := range data{
		ret := ActivityList{}
		ret.Id  = v.Id
		ret.Etype = v.Etype
		ret.FPostion = v.FPostion
		ret.ResourceId = v.ResourceId
		ret.LinkUrl = v.LinkUrl
		ret.SourcePage = v.SourcePage
		ret.StartTm = v.StartTm
		ret.EndTm = v.EndTm
		ret.IsShow = v.IsShow
		ret.Ctime = v.Ctime
		ret.Utime      = v.Utime

		result,_ :=models.GetHashNameByResourceId(v.ResourceId)
		ret.PicUrl = fmt.Sprintf("%s/%s", beego.AppConfig.String("ad_cdn_url"), result.HashName)
		res = append(res, ret)


	}
	return res, nil
}

func GetPopBg() ([]ActivityList, error) {
	data, err := models.GetAllByEtype(0)
	if err != nil && err.Error() != types.EmptyOrmStr {
		logs.Error("[GetFloating] models GetOneByEtypeAndPostion err :%v ", err)
		return nil, err
	}

	res:= make([]ActivityList,0)
	for _, v := range data{
		ret := ActivityList{}
		ret.Id  = v.Id
		ret.Etype = v.Etype
		ret.FPostion = v.FPostion
		ret.ResourceId = v.ResourceId
		ret.LinkUrl = v.LinkUrl
		ret.SourcePage = v.SourcePage
		ret.StartTm = v.StartTm
		ret.EndTm = v.EndTm
		ret.IsShow = v.IsShow
		ret.Ctime = v.Ctime
		ret.Utime      = v.Utime

		result,_ :=models.GetHashNameByResourceId(v.ResourceId)
		ret.PicUrl = fmt.Sprintf("%s/%s", beego.AppConfig.String("ad_cdn_url"), result.HashName)
		res = append(res, ret)


	}
	return res, nil
}
