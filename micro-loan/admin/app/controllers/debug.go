package controllers

import (
	"bufio"
	"fmt"

	"time"

	"io"
	"micro-loan/common/dao"
	"micro-loan/common/models"
	"micro-loan/common/service"
	"micro-loan/common/thirdparty/advance"
	"os"

	"github.com/astaxie/beego/orm"
	//"micro-loan/common/models"
	//"micro-loan/common/types"
	//"micro-loan/common/thirdparty/advance"
	"micro-loan/common/lib/gaws"
)

type DebugController struct {
	BaseController
}

func (c *DebugController) Prepare() {
	// 调用上一级的 Prepare 方法
	c.BaseController.Prepare()

	c.Data["Controller"] = "debug"
}

const (
	READFILE = "/tmp/20180723-184244-f226/test.txt"
	WRITFILE = "/tmp/a.txt"
	TESTFILE = "/tmp/90.txt"
)

func (c *DebugController) Get() {
	c.Data["Action"] = "index"
	timeNow := time.Now()
	c.Data["ServerTime"] = fmt.Sprintf("%v", timeNow)

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
		PicAdvanceOne(acount_id, TESTFILE)
	}
	c.Layout = "layout.html"
	c.TplName = "debug.tpl"
}

func PicAdvanceOne(account_id int64, path string) {
	//身份证照片
	accountProfile, _ := dao.GetAccountProfile(account_id)
	IdPhotoResource, _ := service.OneResource(accountProfile.IdPhoto)

	idPhotoTmp := gaws.BuildTmpFilename(accountProfile.IdPhoto)
	_, err := gaws.AwsDownload(IdPhotoResource.HashName, idPhotoTmp)
	fmt.Println("err----------------------~: ", err)

	//获取live_verify
	livingModel, _ := dao.CustomerLiveVerify(account_id)
	livingPhotoTmp := gaws.BuildTmpFilename(livingModel.ImageEnv)
	livingResource, _ := service.OneResource(livingModel.ImageEnv)
	_, err1 := gaws.AwsDownload(livingResource.HashName, livingPhotoTmp)
	fmt.Println("err1----------------------~: ", err1)

	simily, err := advance.FaceComparison(account_id, idPhotoTmp, livingPhotoTmp)
	fmt.Println("-------------", simily, err)
	lineStr := fmt.Sprintf("%v_%v_%v_%s:%v", account_id, accountProfile.IdPhoto, livingModel.ImageEnv, "匹配率", simily)
	WrittoFile(path, lineStr)
}

func WrittoFile(path, lineStr string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		fmt.Printf("create map file error: %v\n", err)
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
		fmt.Printf("Error: %s\n", err)
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
		fmt.Println("***************************************a :", string(a))
		res = append(res, string(a))
	}
	return res
}

//根据制定范围获取相应的account_id
func AccountProfileAll(a, b int) (list []models.AccountProfile, e error) {
	var obj = models.AccountProfile{}
	o := orm.NewOrm()
	o.Using(obj.Using())

	//where := fmt.Sprintf("%v%v%v%v", " WHERE face_comparison > ", a, " AND face_comparison <", b)
	//sqlOrder := "SELECT account_id, face_comparison  from " + models.ACCOUNT_PROFILE_TABLENAME + where
	sqlOrder := "SELECT account_id, face_comparison  from " + models.ACCOUNT_PROFILE_TABLENAME
	// sqlOrder = fmt.Sprintf("%s%d", sqlOrder)
	o.Raw(sqlOrder).QueryRows(&list)
	fmt.Println("--------", e)
	return
}
