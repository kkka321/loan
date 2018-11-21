package main

import (
	//"encoding/json"
	"fmt"
	"io/ioutil"

	_ "micro-loan/common/lib/clogs"
	_ "micro-loan/common/lib/db/mysql"

	"micro-loan/common/service"
	//"micro-loan/common/thirdparty/advance"
	//"micro-loan/common/thirdparty/faceid"
	//"micro-loan/common/thirdparty/textlocal"
	"micro-loan/common/tools"

	"github.com/astaxie/beego/logs"
)

func main() {
	testAddSlashes()
}

func testAddSlashes() {
	str := `Is your \ name O'reilly?"~~~ \\\`
	strA := tools.AddSlashes(str)
	fmt.Printf("AddSlashes(`%s`): %s\n", str, strA)
	strB := tools.StripSlashes(strA)
	fmt.Printf("StripSlashes(`%s`): %s\n", strA, strB)
}

func testMain() {
	logs.Debug("debug api ...")

	url := "http://localhost/post.php"
	queryString := map[string]string{
		"action": "upload",
		"env":    "dev",
	}
	reqHeaders := map[string]string{
		"Connection": "keep-alive",
		"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.132 Safari/537.36",
	}
	files := map[string]string{
		"file1": "./README.md",
		"file2": "./debug.go",
	}

	resByte, httpCode, err := tools.MultipartClient(url, queryString, reqHeaders, files, tools.DefaultHttpTimeout())
	fmt.Printf("resByte: %s, httpCode: %d, err: %v\n", resByte, httpCode, err)

	uuid := tools.GetGuid()
	fmt.Printf("uuid: %v\n", uuid)

	inputFile := "./debug.go"

	hashDir, hashName, fileMd5, err := tools.BuildFileHashName(inputFile)
	fmt.Printf("BuildFileHashName ->       hashDir: %s, hashName: %s, err: %v\n", hashDir, hashName, err)

	buf, _ := ioutil.ReadFile(inputFile)
	hashDir, hashName, fileMd5 = tools.BuildUploadFileHashName(buf, "go")
	fmt.Printf("BuildUploadFileHashName -> hashDir: %s, hashName: %s\n", hashDir, hashName)
	fmt.Printf("md5sum(%s): %s\n", inputFile, fileMd5)

	localHashDir := tools.LocalHashDir(hashDir)
	fmt.Printf("localHashDir: %s\n", localHashDir)

	binFile := "/opt/data/Indonesia/11.jpeg"
	extension, mime, err := tools.DetectFileType(binFile)
	fmt.Printf("binFile: %s, extension: %s, mime: %s, err: %v\n", binFile, extension, mime, err)

	err = tools.Remove("./nofile")
	fmt.Printf("rm ./nofile err: %v\n", err)

	//// 调试advance
	//param := map[string]interface{}{}
	//file := map[string]interface{}{
	//	"ocrImage": "/opt/data/Indonesia/1.jpeg",
	//}
	//resByte, resData, err := advance.Request(54321, advance.ApiOCR, param, file)
	//fmt.Printf("resByte: %s, resData: %v, err: %v, IsSuccess: %v\n", resByte, resData, err, advance.IsSuccess(resData.Code))
	//dataMap := resData.Data.(map[string]interface{})
	//fmt.Printf("dataMap: %v\n", dataMap)

	//// 调试 faceid
	//faceFiles := map[string]string{
	//	"image_best": "/opt/data/faceid/image_best.jpeg",
	//	"image_env":  "/opt/data/faceid/image_env.jpeg",
	//	"image_ref1": "/opt/data/faceid/image_action1.jpeg",
	//	"image_ref2": "/opt/data/faceid/image_action2.jpeg",
	//	"image_ref3": "/opt/data/faceid/image_action3.jpeg",
	//}

	//delta := "abctest"
	//originRes, httpCode, err := faceid.Verify(67890, faceid.ComparisonTypeDefault, faceid.FaceImageTypeDefault, faceFiles, delta)
	//fmt.Printf("originRes: %s, httpCode: %d, err: %v\n", originRes, httpCode, err)

	esRes, _, _, err := service.EsSearchById("0f6148f4b21f3ed7097b3b0dcc790e35")
	fmt.Printf("esRes: %#v, err: %v\n", esRes, err)

	//age, err := service.CustomerAge("3172026605991002")
	age, err := service.CustomerAge("3173024402000007")
	fmt.Printf("age: %d, err: %v\n", age, err)

	service.OrderListLoanStatus4Review()

	nt := tools.NaturalDay(1)
	fmt.Printf("nt: %d, date: %s\n", nt, tools.MDateMHS(nt))

	query := "a=just like it"
	fmt.Printf(`urlencode("%s") = %s%s`, query, tools.UrlEncode(query), "\n")
	fmt.Printf("rawurlencode(%s = %s\n", query, tools.RawUrlEncode(query))

	//numbers := []string{
	//	"8861429589",
	//}
	//message := "Just test send sms message. no 0091"
	//apiRes, err := textlocal.SendSms(numbers, message, 0, textlocal.SenderDefault, textlocal.TestYes)
	//fmt.Printf("apiRes: %v, err: %v\n", apiRes, err)

	var hit int
	var max = 100000
	for i := 0; i < max; i++ {
		r := tools.GenerateRandom(1, 101)
		if r > 99 {
			hit++
		}
	}
	fmt.Printf("max: %d, hit: %d, rate: %.4f\n", max, hit, float64(hit)/float64(max))

	originName := "MELI AMALIA MUTAQIN,S.  SOS T  -  T, .  !_O"
	name := tools.TrimRealName(originName)
	fmt.Printf("originName: %s\n", originName)
	fmt.Printf("      name: %s\n", name)
}
