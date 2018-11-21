package service

import (
	"fmt"
	"micro-loan/common/models"
	"micro-loan/common/tools"
	"micro-loan/common/types"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
)

type AdResponse struct {
	AdUrl      string
	LinkUrl    string
	SourcePage int64
}

func GetAdvertisement() (map[string]interface{}, error) {
	res, err := models.OneAdvertisementByTm()
	if err != nil && err.Error() != types.EmptyOrmStr {
		logs.Error("[GetAdvertisement] models OneAdvertisementByTm err :%v ", err)
		return nil, err
	}

	aDUrl := ""
	if res != (models.Advertisement{}) {
		pic, err := OneResource(res.ResourceId)
		if err != nil {
			logs.Error("[GetAdvertisement] models OneResource err :%v ", err)
			return nil, err
		}

		aDUrl = fmt.Sprintf("%s/%s", beego.AppConfig.String("ad_cdn_url"), pic.HashName)

	} else {
		res = models.Advertisement{}
	}

	is_ad := false
	if res.ResourceId > 0 {
		is_ad = true
	}

	mp := map[string]interface{}{
		"is_ad":       is_ad,
		"ad_url":      aDUrl,
		"link_url":    res.LinkUrl,
		"source_page": res.SourcePage,
		"server_time": tools.GetUnixMillis(),
	}
	return mp, nil
}

type AdvertisementList struct {
	Id         int64
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
func GetMultiAdvertisements() (datas []AdvertisementList, err error) {
	data, err := models.GetMultiAdvertisements()
	if err != nil && err.Error() != types.EmptyOrmStr {
		logs.Error("[GetMultiAdvertisements] models GetMultiAdvertisements err :%v ", err)
		return nil, err
	}
	if len(data) == 0 {
		datas = []AdvertisementList{}
		return datas, nil
	}


	for _, v:=range data {
		ret := AdvertisementList{}

		ret.Id = v.Id
		ret.ResourceId = v.ResourceId
		ret.LinkUrl = v.LinkUrl
		ret.SourcePage = v.SourcePage
		ret.StartTm = v.StartTm
		ret.EndTm = v.EndTm
		ret.IsShow = v.IsShow
		ret.Ctime = v.Ctime
		ret.Utime = v.Utime
		result,_ :=models.GetHashNameByResourceId(v.ResourceId)
		ret.PicUrl = fmt.Sprintf("%s/%s", beego.AppConfig.String("ad_cdn_url"), result.HashName)
		datas= append(datas, ret)
	}

	return
}
