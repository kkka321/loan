package controllers

import (
	"micro-loan/common/cerror"
	"micro-loan/common/dao"
	"micro-loan/common/lib/gaws"
	"micro-loan/common/service"
	"micro-loan/common/thirdparty/advance"
	"micro-loan/common/tools"
	"micro-loan/common/types"
)

type ReLoanController struct {
	ApiBaseController
}

func (c *ReLoanController) Prepare() {
	// 调用上一级的 Prepare 方
	c.ApiBaseController.Prepare()

	// 统一将 ip 加到 RequestJSON 中
	c.RequestJSON["ip"] = c.Ctx.Input.IP()
	c.RequestJSON["related_id"] = int64(0)
}

func (c *ReLoanController) ReLoanUploadHandHeldInPhoto() {
	// 查看是否可以复贷
	/*
		if !service.IsRepeatLoan(c.AccountID) {
			c.Data["json"] = cerror.BuildApiResponse(cerror.MismatchRepeatLoan, "")
			c.ServeJSON()
			return
		}


			order, _ := service.AccountLastLoanOrder(c.AccountID)

			if order.CheckStatus != types.LoanStatusInvalid && order.CheckStatus != types.LoanStatusAlreadyCleared && order.CheckStatus != types.LoanStatusReject {
				//有进行中的订单，不符合复贷
				c.Data["json"] = cerror.BuildApiResponse(cerror.OriginPendingOrder, "")
				c.ServeJSON()
				return
			}

			if order.CheckStatus == types.LoanStatusReject {
				today := tools.NaturalDay(0)
				futureValidDay := tools.BaseDayOffset(order.ApplyTime, 7)
				if today <= futureValidDay {
					//被拒7天内，也不符合复贷
					c.Data["json"] = cerror.BuildApiResponse(cerror.OriginOrder7DaysReject, "")
					c.ServeJSON()
					return
				}
			}
	*/

	//检测之前是存在手持照片
	if !service.IfOriginHandHeldIdPhoneExist(c.AccountID) {
		c.Data["json"] = cerror.BuildApiResponse(cerror.OriginHandHeldIdPhotoNotExist, "")
		c.ServeJSON()
		return
	}
	//复贷时,上传新的手持照片
	resourceId, handHeldIdPhotoTmp, code, err := c.UploadResource("fs1", types.Use2ReLoanHandHeldIdPhoto)
	defer tools.Remove(handHeldIdPhotoTmp)
	if err != nil {
		c.Data["json"] = cerror.BuildApiResponse(code, "")
		c.ServeJSON()
		return
	}

	accountProfile, _ := dao.GetAccountProfile(c.AccountID)
	handIdPhotoResource, _ := service.OneResource(accountProfile.HandHeldIdPhoto)

	idPhotoTmp := gaws.BuildTmpFilename(accountProfile.HandHeldIdPhoto)
	gaws.AwsDownload(handIdPhotoResource.HashName, idPhotoTmp)
	//方法执行完 删除tmp下的图片
	defer tools.Remove(idPhotoTmp)

	// 3. 如果有人脸,调用advance.ai,进行比对,并回写数据
	//// 3.1 Face Comparison 人脸比对
	similarity, _ := advance.FaceComparison(c.AccountID, idPhotoTmp, handHeldIdPhotoTmp)

	data := map[string]interface{}{}
	data["server_time"] = tools.GetUnixMillis()
	data["result"] = false

	if similarity >= 70 {
		accountProfile.SaveFaceComparison(similarity)
		service.AddReLoanImage(c.AccountID, resourceId)
		service.UploadReLoanPhotoSuccess(c.AccountID)
		data["result"] = true
		data["current_step"] = service.ProfileCompletePhase(c.AccountID, c.UIVersion, c.VersionCode)
		data["similarity"] = similarity
		c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, data)
	} else {
		data["current_step"] = service.ProfileCompletePhase(c.AccountID, c.UIVersion, c.VersionCode)
		data["similarity"] = similarity
		c.Data["json"] = cerror.BuildApiResponse(cerror.TwoImagesCompareError, data)
	}
	c.ServeJSON()
}
