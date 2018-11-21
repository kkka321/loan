package service

import (
	"fmt"
	"micro-loan/common/models"
	"micro-loan/common/tools"
	"micro-loan/common/types"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
)

func GetBanners() ([]map[string]interface{}, error) {
	res, err := models.GetMultiBannersByType(types.BannerTypeHomePage)
	if err != nil && err.Error() != types.EmptyOrmStr {
		return []map[string]interface{}{}, err
	}

	mps := make([]map[string]interface{}, 0)
	for _, v := range res {
		pic, err := OneResource(v.ResourceId)
		if err != nil {
			logs.Error("[GetAdvertisement] models OneResource err :%v ", err)
			return nil, err
		}

		picUrl := fmt.Sprintf("%s/%s", beego.AppConfig.String("ad_cdn_url"), pic.HashName)

		mp := map[string]interface{}{
			"postion":     v.Postion,    //banner展示的位置
			"pic_url":     picUrl,       //banner图片地址
			"link_url":    v.LinkUrl,    //链接
			"source_page": v.SourcePage, // 大于0跳原生页面,小于等于0跳h5页面
		}

		mps = append(mps, mp)
	}
	return mps, nil
}

func GetInvite() (invite map[string]interface{}, err error) {
	res, err := models.GetMultiBannersByType(types.BannerTypeInvitePage)
	if err != nil && err.Error() != types.EmptyOrmStr {
		return
	}

	for _, v := range res {
		if v.Id <= 0 {
			continue
		}

		t := tools.GetUnixMillis()
		if t < v.StartTime || t > v.EndTime {
			continue
		}

		pic, err1 := OneResource(v.ResourceId)
		if err1 != nil {
			err = err1
			logs.Info("[GetInvite] models OneResource err :%v ", err)
			continue
		}

		picUrl := fmt.Sprintf("%s/%s", beego.AppConfig.String("ad_cdn_url"), pic.HashName)

		invite = map[string]interface{}{
			"pic_url":       picUrl,        //邀请好友页图片地址
			"text":          v.Content,     //图片上的文字
			"font_color":    v.FontColor,   //图片上的文字颜色
			"font_link_url": v.FontLinkUrl, //图片上文字的跳转链接
		}

		break
	}

	return
}

type BannerList struct {
	Id          int64
	ResourceId  int64
	LinkUrl     string
	SourcePage  int64
	Postion     int64
	BannerType  int
	StartTime   int64
	EndTime     int64
	Content     string
	FontColor   string
	FontLinkUrl string
	Ctime       int64
	Utime       int64
	PicUrl      string
}

func GetMultiBanners() (datas []BannerList, err error) {
	data, err := models.GetMultiBanners()
	if err != nil && err.Error() != types.EmptyOrmStr {
		logs.Error("[GetMultiBanners] models GetMultiBanners err :%v ", err)
		return nil, err
	}
	if len(data) == 0 {
		datas = []BannerList{}
		return datas, nil
	}

	for _, v := range data {
		ret := BannerList{}

		ret.Id = v.Id
		ret.ResourceId = v.ResourceId
		ret.LinkUrl = v.LinkUrl
		ret.SourcePage = v.SourcePage
		ret.Postion = v.Postion
		ret.BannerType = v.BannerType
		ret.StartTime = v.StartTime
		ret.EndTime = v.EndTime
		ret.Content = v.Content
		ret.FontColor = v.FontColor
		ret.FontLinkUrl = v.FontLinkUrl
		ret.Ctime = v.Ctime
		ret.Utime = v.Utime

		result, _ := models.GetHashNameByResourceId(v.ResourceId)
		ret.PicUrl = fmt.Sprintf("%s/%s", beego.AppConfig.String("ad_cdn_url"), result.HashName)
		datas = append(datas, ret)
	}

	return
}
