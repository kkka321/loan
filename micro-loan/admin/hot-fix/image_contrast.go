package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"micro-loan/common/service"
	"micro-loan/common/tools"
	"strings"

	"github.com/astaxie/beego/logs"

	"io"
	"micro-loan/common/dao"
	"micro-loan/common/models"
	"micro-loan/common/thirdparty/advance"
	"os"

	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"

	"github.com/astaxie/beego/orm"
	//"micro-loan/common/models"
	//"micro-loan/common/types"
	//"micro-loan/common/thirdparty/advance"
)

const (
	READFILE   = "/tmp/20180723-184244-f226/test.txt"
	WRITFILE   = "/tmp/a.txt"
	TESTFILE10 = "/tmp/10_20_compare.txt"
	TESTFILE20 = "/tmp/20_35_compare.txt"
	TESTFILE35 = "/tmp/35_50_compare.txt"
	TESTFILE50 = "/tmp/50_70_compare.txt"
	TESTFILE   = "/tmp/b.txt"
)

func main() {

	//读取account_id 获取id_check 和　faceid
	// res := []int64{
	// 	180319010000078023,
	// 	180316010000012137,
	// 	180321010000605613,
	// 	180322010000718302,
	// 	180323010000942984,
	// 	180323010000967391,
	// 	180323010001152790,
	// 	180323010001303256,
	// 	180323010001565122,
	// 	180324010001805402,
	// 	180324010002715166,
	// 	180324010002770548,
	// 	180324010002923176,
	// 	180324010003174957,
	// 	180324010004242525,
	// 	180325010004412550,
	// 	180325010004577211,
	// 	180325010004809884,
	// 	180326010005804613,
	// 	180326010005980667,
	// 	180326010005689581,
	// 	180325010005390852,
	// 	180326010006439145,
	// 	180327010006696938,
	// 	180327010006911741,
	// 	180326010006484644,
	// 	180327010007108544,
	// 	180327010007226203,
	// 	180327010007332984,
	// 	180327010007507882,
	// 	180327010007800879,
	// 	180327010008155024,
	// 	180328010009335615,
	// 	180328010009559500,
	// 	180328010009724170,
	// 	180329010009918240,
	// 	180329010009969015,
	// 	180329010010060598,
	// 	180329010010241855,
	// 	180329010010122810,
	// 	180329010010822329,
	// 	180329010011298741,
	// 	180329010011253384,
	// 	180329010011500400,
	// 	180329010011558414,
	// 	180329010011879362,
	// 	180329010012274908,
	// 	180330010012535150,
	// 	180330010012760695,
	// 	180330010012735337,
	// 	180330010013061205,
	// 	180330010013520426,
	// 	180330010013715182,
	// 	180330010014319045,
	// 	180330010014418390,
	// 	180330010015024331,
	// 	180330010015773131,
	// 	180330010016007770,
	// 	180331010016769695,
	// 	180331010016817152,
	// 	180330010015749985,
	// 	180331010017262770,
	// 	180331010017593997,
	// 	180331010017769711,
	// 	180331010017537725,
	// 	180331010016232094,
	// 	180331010018282867,
	// 	180331010018350629,
	// 	180331010017994030,
	// 	180331010020356495,
	// 	180331010020555840,
	// 	180331010020934063,
	// 	180331010021033113,
	// 	180331010021675036,
	// 	180331010021772164,
	// 	180331010021897231,
	// 	180331010022027098,
	// 	180331010022228156,
	// 	180331010022275788,
	// 	180331010022535245,
	// 	180331010022634347,
	// 	180331010023525905,
	// 	180401010025789262,
	// 	180401010026295007,
	// 	180401010026796099,
	// 	180401010026960797,
	// 	180401010027067200,
	// 	180401010027525366,
	// 	180401010028931617,
	// 	180401010029106202,
	// 	180401010029254393,
	// 	180401010029780911,
	// 	180331010020853823,
	// 	180401010030285099,
	// 	180401010029719851,
	// 	180401010030837506,
	// 	180401010031337247,
	// 	180401010031559933,
	// 	180402010031968485,
	// 	180402010032095441,
	// 	180708010387422772,
	// 	180717015658536057,
	// 	180717015700414177,
	// 	180715014583677124,
	// 	180717015710322016,
	// 	180717015713236604,
	// 	180717015714151230,
	// 	180716015464782749,
	// 	180717015716612583,
	// 	180717015721188909,
	// 	180717015724399489,
	// 	180717015728609660,
	// 	180717015728271982,
	// 	180717015728863846,
	// 	180717015730793603,
	// 	180717015731217127,
	// 	180717015734687694,
	// 	180704018082576149,
	// 	180717015742967160,
	// 	180605010925938580,
	// 	180716015325345542,
	// 	180717015669926507,
	// 	180717015678048872,
	// 	180717015677278953,
	// 	180717015716612583,
	// 	180716015365344814,
	// 	180531016314138783,
	// 	180628013440706500,
	// 	180717015726332716,
	// 	180717015733076295,
	// 	180717015705733563,
	// 	180717015749838029,
	// 	180717015751707314,
	// 	180717015752314032,
	// 	180717015658068624,
	// 	180717015611716546,
	// 	180717015754846194,
	// 	180717015675145877,
	// 	180717015759279262,
	// 	180717015760408338,
	// 	180603018559528000,
	// 	180717015632147947,
	// 	180717015702071528,
	// 	180717015709960039,
	// 	180712012611435328,
	// 	180717015753729570,
	// 	180716015454678429,
	// 	180717015771126362,
	// 	180707019651584046,
	// 	180717015773241333,
	// }
	res := []int64{
		180603018608047272,
		180720017345850524,
		180525014787171156,
		180702016385837066,
		180602017984591489,
		180711011874302034,
		180719016848593509,
		180610014927495841,
		180531016258214461,
		180718016335759778,
		180707019557452847,
		180630014904185805,
		180605010731943597,
		180614017303653066,
		180426012284154670,
		180602017822170402,
		180603018898882172,
		180622019941796836,
		180531016298555364,
		180718015956946350,
		180620018649145754,
		180611015483190180,
		180702016369510669,
		180717015745202021,
		180420010906830978,
		180609014233110895,
		180426012322551653,
		180717015694569295,
		180605011017921726,
		180531016060720779,
		180416020530488758,
		180529025536435945,
		180601027188407507,
		180601027343519982,
		180603028802842392,
		180605020753895322,
		180605021098578607,
		180606022076609700,
		180606022177431077,
		180607022623354364,
	}
	for _, acount_id := range res {
		PicAdvanceOnes(acount_id, TESTFILE)
	}
	//AccountProfileAll()
}

type Dates struct {
	Similarity float64 `json:"similarity"`
}

type Third struct {
	Code    string `json:"code"`
	Data    Dates
	Extra   string `json:"extra"`
	Message string `json:"message"`
}

type Fil struct {
	IdHoldingImage string `json:"idHoldingImage"`
	QueryString    string `json:"query_string"`
}
type Files struct {
	Files Fil
}

/*



 */

//根据制定范围获取相应的account_id
func AccountProfileAll() (e error) {
	var obj = models.ThirdpartyRecord{}
	o := orm.NewOrm()
	o.Using(obj.Using())
	sqls := "SELECT api, related_id,response, request from " + models.THIRDPARTY_RECORD_TABLENAME
	thirres := []models.ThirdpartyRecord{}
	o.Raw(sqls).QueryRows(&thirres)
	var third Third
	var files Files
	var a, b, c, d int
	for _, v := range thirres {
		num := a + b + c + d
		if num == 600 {
			break
		}
		flag := (v.Api == "https://api.advance.ai/openapi/face-recognition/v2/id-check")

		if !flag {
			continue
		}

		if err := json.Unmarshal([]byte(v.Response), &third); err == nil {

			if err := json.Unmarshal([]byte(v.Request), &files); err == nil {
				var objs = models.UploadResource{}
				o := orm.NewOrm()
				o.Using(objs.Using())

				tmp := strings.Replace(files.Files.IdHoldingImage, "/tmp/", "", -1)
				tmp = strings.Replace(tmp, ".", "", -1)
				upload := models.UploadResource{}
				err := o.QueryTable(objs.TableName()).Filter("content_md5", tmp).One(&upload)
				if err != nil {
					fmt.Println("**********************err:", err)
				}
				fmt.Println("**********************hash:", upload.HashName)
				if third.Data.Similarity >= 10 && third.Data.Similarity <= 20 && a < 20 {

					accountProfile, _ := dao.GetAccountProfile(v.RelatedId)
					IdPhotoResource, _ := service.OneResource(accountProfile.IdPhoto)

					idPhotoTmp := service.BuildTmpFilename(v.RelatedId)
					_, err := service.AwsDownload(IdPhotoResource.HashName, idPhotoTmp)
					if err == nil {

						tmpStr := fmt.Sprintf("%v_%v_%v", v.RelatedId, "匹配率", third.Data.Similarity)
						WrittoFile(TESTFILE, tmpStr)
						lineStr := fmt.Sprintf("%v_%v_%v_%s: %v", v.RelatedId, tmp, upload.HashName, "匹配率", third.Data.Similarity)
						WrittoFile(TESTFILE10, lineStr)

						a++
					} else {
						fmt.Println("err----------:", err)
						fmt.Println("HashName----------:", upload.HashName)
					}

				} else if third.Data.Similarity > 20 && third.Data.Similarity <= 35 && b < 40 {

					accountProfile, _ := dao.GetAccountProfile(v.RelatedId)
					IdPhotoResource, _ := service.OneResource(accountProfile.IdPhoto)

					idPhotoTmp := service.BuildTmpFilename(v.RelatedId)
					_, err := service.AwsDownload(IdPhotoResource.HashName, idPhotoTmp)
					if err == nil {

						tmpStr := fmt.Sprintf("%v_%v_%v", v.RelatedId, "匹配率", third.Data.Similarity)
						WrittoFile(TESTFILE, tmpStr)
						lineStr := fmt.Sprintf("%v_%v_%v_%s: %v", v.RelatedId, tmp, upload.HashName, "匹配率", third.Data.Similarity)
						WrittoFile(TESTFILE20, lineStr)

						b++
					} else {
						fmt.Println("err----------:", err)
						fmt.Println("HashName----------:", upload.HashName)
					}

				} else if third.Data.Similarity > 35 && third.Data.Similarity <= 50 && c < 40 {

					accountProfile, _ := dao.GetAccountProfile(v.RelatedId)
					IdPhotoResource, _ := service.OneResource(accountProfile.IdPhoto)

					idPhotoTmp := service.BuildTmpFilename(v.RelatedId)
					_, err := service.AwsDownload(IdPhotoResource.HashName, idPhotoTmp)
					if err == nil {

						tmpStr := fmt.Sprintf("%v_%v_%v", v.RelatedId, "匹配率", third.Data.Similarity)
						WrittoFile(TESTFILE, tmpStr)
						lineStr := fmt.Sprintf("%v_%v_%v_%s: %v", v.RelatedId, tmp, upload.HashName, "匹配率", third.Data.Similarity)
						WrittoFile(TESTFILE35, lineStr)

						c++
					} else {
						fmt.Println("err----------:", err)
						fmt.Println("HashName----------:", upload.HashName)
					}

				} else if third.Data.Similarity > 50 && third.Data.Similarity <= 70 && d < 500 {

					accountProfile, _ := dao.GetAccountProfile(v.RelatedId)
					IdPhotoResource, _ := service.OneResource(accountProfile.IdPhoto)

					idPhotoTmp := service.BuildTmpFilename(v.RelatedId)
					_, err := service.AwsDownload(IdPhotoResource.HashName, idPhotoTmp)
					if err == nil {

						tmpStr := fmt.Sprintf("%v_%v_%v", v.RelatedId, "匹配率", third.Data.Similarity)
						WrittoFile(TESTFILE, tmpStr)
						lineStr := fmt.Sprintf("%v_%v_%v_%s: %v", v.RelatedId, tmp, upload.HashName, "匹配率", third.Data.Similarity)
						WrittoFile(TESTFILE50, lineStr)

						d++
					} else {
						fmt.Println("err----------:", err)
						fmt.Println("HashName----------:", upload.HashName)
					}

				}

			}
		}
	}

	// var obj = models.AccountProfile{}
	// o := orm.NewOrm()
	// o.Using(obj.Using())

	// where := fmt.Sprintf("%v%v%v%v%v%v", " WHERE face_comparison > ", a, " AND face_comparison <", b, " limit ", limit)
	// sqlOrder := "SELECT account_id, face_comparison  from " + models.ACCOUNT_PROFILE_TABLENAME + where
	// //sqlOrder := "SELECT account_id, face_comparison  from " + models.ACCOUNT_PROFILE_TABLENAME
	// // sqlOrder = fmt.Sprintf("%s%d", sqlOrder)
	// o.Raw(sqlOrder).QueryRows(&list)
	// fmt.Println("--------", e)
	return nil
}

func PicAdvanceOne(account_id int64, path string) {
	//身份证照片
	idPhotoTmp := service.BuildTmpFilename(account_id)

	//获取handid
	accountProfile, _ := dao.GetAccountProfile(account_id)
	handIdPhotoResource, _ := service.OneResource(accountProfile.HandHeldIdPhoto)

	handPhotoTmp := service.BuildTmpFilename(accountProfile.HandHeldIdPhoto)
	_, err1 := service.AwsDownload(handIdPhotoResource.HashName, handPhotoTmp)
	logs.Error("err1----------------------~: ", err1)

	simily, err := advance.FaceComparison(account_id, idPhotoTmp, handPhotoTmp)
	logs.Error("-------------", simily, err)
	lineStr := fmt.Sprintf("%v_%v_%v_%s: %v", account_id, idPhotoTmp, handPhotoTmp, "匹配率", simily)
	WrittoFile(path, lineStr)
}

func WrittoFile(path, lineStr string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		logs.Error("create map file error: %v\n", err)
		return err
	}
	defer f.Close()

	n, err := f.Write([]byte(lineStr + "\n"))
	if err != nil {
		fmt.Println(err, n)
	}

	return nil
}

func ReadtoFile(path string) []string {
	fi, err := os.Open(path)
	if err != nil {
		logs.Error("Error: %s\n", err)
		return []string{}
	}
	defer fi.Close()
	res := []string{}
	br := bufio.NewReader(fi)
	for {
		a, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}
		logs.Error("***************************************a :", string(a))
		res = append(res, string(a))
	}
	return res
}

func PicAdvanceOnes(account_id int64, path string) {
	//身份证照片
	accountProfile, _ := dao.GetAccountProfile(account_id)
	IdPhotoResource, _ := service.OneResource(accountProfile.IdPhoto)

	idPhotoTmp := service.BuildTmpFilename(accountProfile.IdPhoto)
	_, err := service.AwsDownload(IdPhotoResource.HashName, idPhotoTmp)
	logs.Error("err----------------------~: ", err)
	defer tools.Remove(idPhotoTmp)
	//获取live_verify
	livingModel, _ := dao.CustomerLiveVerify(account_id)
	livingPhotoTmp := service.BuildTmpFilename(livingModel.ImageEnv)
	livingResource, _ := service.OneResource(livingModel.ImageEnv)
	_, err1 := service.AwsDownload(livingResource.HashName, livingPhotoTmp)
	logs.Error("err1----------------------~: ", err1)
	defer tools.Remove(livingPhotoTmp)

	simily, err := advance.FaceComparison(account_id, idPhotoTmp, livingPhotoTmp)
	logs.Error("-------------", simily, err)
	lineStr := fmt.Sprintf("%v_%v_%v_%s: %v", account_id, accountProfile.IdPhoto, livingModel.ImageEnv, "匹配率", simily)
	WrittoFile(path, lineStr)
}
