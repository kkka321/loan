package service

import (
	"fmt"
	"micro-loan/common/dao"
	"micro-loan/common/models"
	"micro-loan/common/pkg/system/config"
	"micro-loan/common/tools"
	"micro-loan/common/types"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
)

func isDisplayAd(accountId int64) (isDisplay bool, companyId int) {
	order, err := dao.AccountLastLoanOrder(accountId)
	if err != nil {
		return
	}

	// 首贷的反欺诈拒绝/黑名单拒绝, 显示广告位
	if order.IsReloan == 0 && (order.RiskCtlStatus == types.RiskCtlAFReject || order.RiskCtlStatus == types.RiskCtlThirdBlacklistReject) {

		/* 根据配置获取广告位显示的广告 */
		// 流量互换合作公司数量
		num, _ := config.ValidItemInt("traffic_exchange_company_num")
		// 流量互换合作公司比例分配
		percentage := config.ValidItemString("traffic_exchange_percentage")
		percentageArr := strings.Split(percentage, ",")
		if num > 0 && len(percentageArr) > 0 {
			var percentageIntArr []int

			for _, v := range percentageArr {
				perInt, _ := tools.Str2Int(v)
				percentageIntArr = append(percentageIntArr, perInt)
			}

			randomValue := tools.GenerateRandom(1, 101)
			var sum int
			for k, v := range percentageIntArr {
				// 按照合作公司数量限制比例配置
				if num <= k {
					break
				}

				sum += v
				if sum >= randomValue {
					isDisplay = true
					companyId = k + 1

					break
				}
			}
		}
	}

	return
}

func GetAdPositionDisplay(accountId int64, position int) (data map[string]interface{}, err error) {
	var picUrl string
	var pic models.UploadResource
	var res models.AdPosition

	is_ad, companyId := isDisplayAd(accountId)
	if !is_ad {
		goto next
	}

	res, err = models.GetAdPositionByCompanyIdAndPosition(companyId, position)
	if err != nil && err.Error() != types.EmptyOrmStr {
		return
	}

	if res.ResourceId <= 0 {
		goto next
	}

	pic, err = OneResource(res.ResourceId)
	if err != nil {
		logs.Error("[GetAdPosition] models OneResource err :%v ", err)
		return
	}

	picUrl = fmt.Sprintf("%s/%s", beego.AppConfig.String("ad_cdn_url"), pic.HashName)

next:
	data = map[string]interface{}{
		"is_ad":       is_ad,
		"ad_pic_url":  picUrl,      //广告位图片地址
		"ad_link_url": res.LinkUrl, //链接
	}

	return
}

func ApiDataAddAdPosition(adPosition, data map[string]interface{}) {

	data["is_ad"] = adPosition["is_ad"]
	data["ad_pic_url"] = adPosition["ad_pic_url"]
	data["ad_link_url"] = adPosition["ad_link_url"]
}

type AdPositionList struct {
	Id         int64
	ResourceId int64
	PicUrl     string
	LinkUrl    string
	CompanyId  int
	Position   int
	Ctime      int64
	Utime      int64
}

func GetAdPositionList() (datas []AdPositionList, err error) {
	data, err := models.GetMultiAdPosition()
	if err != nil && err.Error() != types.EmptyOrmStr {
		logs.Error("[GetAdPositionList] models GetAdPositionList err :%v ", err)
		return nil, err
	}
	if len(data) == 0 {
		datas = []AdPositionList{}
		return datas, nil
	}

	for _, v := range data {
		ret := AdPositionList{}

		ret.Id = v.Id
		ret.ResourceId = v.ResourceId
		ret.LinkUrl = v.LinkUrl
		ret.CompanyId = v.CompanyId
		ret.Position = v.Position
		ret.Ctime = v.Ctime
		ret.Utime = v.Utime

		result, _ := models.GetHashNameByResourceId(v.ResourceId)
		ret.PicUrl = fmt.Sprintf("%s/%s", beego.AppConfig.String("ad_cdn_url"), result.HashName)
		datas = append(datas, ret)
	}

	return
}
