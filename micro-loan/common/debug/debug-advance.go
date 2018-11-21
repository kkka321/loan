package main

import (
	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"
	"micro-loan/common/thirdparty/advance"
	"micro-loan/common/thirdparty/api253"

	"github.com/astaxie/beego/logs"
)

func main() {
	logs.Debug("debug api ...")
	// IdentiryCheck()
	// FaceComparison()
	IDHoldingPhotoCheck()
	// ocr()
	// FaceCheck()

}

func FaceCheck() {
	accountID := int64(180301010007362546)
	// ocr := "/tmp/testphoto/ocr.jpeg"
	handhold := "/Users/mac/Documents/jiujie3.jpg"
	score, _ := api253.FaceCheck(accountID, handhold)
	logs.Debug("score", score)
}

func ocr() {
	accountID := int64(180301010007362546)
	ocrPhoto := "/Users/mac/Documents/ocr.jpeg"
	param := map[string]interface{}{}
	file := map[string]interface{}{
		"ocrImage": ocrPhoto,
	}
	_, resData, err := advance.Request(accountID, advance.ApiOCR, param, file)
	logs.Debug("resData:", resData)

	if err == nil && advance.IsSuccess(resData.Code) {
		ocrRealname := resData.Data.Name
		identity := resData.Data.IDNumber

		logs.Debug("orcRealName:", ocrRealname)
		logs.Debug("identity:", identity)

	}
}

func IDHoldingPhotoCheck() {
	accountID := int64(180301010007362546)
	ocr := "/Users/mac/Documents/ocr.jpeg"
	handhold := "/Users/mac/Documents/handhold.jpeg"
	// handhold := "/Users/mac/Documents/handhold.jpeg"
	code, _ := advance.IDHoldingPhotoCheck(accountID, handhold, ocr)
	logs.Debug("code", code)
}

func IdentiryCheck() {

	accountID := int64(180301010007362546)
	name := "MIRA AMALIA WULAN"
	idnumber := "3273205803840003"
	resp, _ := advance.IdentiryCheck(accountID, name, idnumber)

	logs.Debug("resp:", resp)

}

func FaceComparison() {
	accountID := int64(180301010007362546)
	// // ocr := "/tmp/testphoto/ocr.jpeg"
	// handhold := "/Users/mac/Documents/handhold.jpeg"
	// livingbest := "/Users/mac/Documents/livingbig.jpeg"
	handhold := "/Users/mac/Documents/yilei.jpg"
	livingbest := "/Users/mac/Documents/yilei.jpg"

	similar, _ := advance.FaceComparison(accountID, livingbest, handhold)

	logs.Debug("similar:", similar)

}
