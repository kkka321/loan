package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"encoding/base64"

	"github.com/astaxie/beego/logs"
)

const (
	HttpMethodGet  string = "GET"
	HttpMethodPOST string = "POST"
)

type HttpTimeout struct {
	DialTimeout           int
	DialKeepAlive         int
	TLSHandshakeTimeout   int
	ResponseHeaderTimeout int
	ExpectContinueTimeout int
	Timeout               int
}

func DefaultHttpTimeout() HttpTimeout {
	return SetHttpTimeout(30, 60, 60, 90, 30, 90)
}

/**
* 根据常用contenType 构造requestHeaders requestBody
* 传参：
* 1.x-www-form-urlencoded
* 2.json
**/
func MakeReqHeadAndBody(ctypeName string, param map[string]interface{}) (requestBody string, requestHeaders map[string]string) {

	var contentType string
	requestHeaders = make(map[string]string)
	switch ctypeName {
	case "x-www-form-urlencoded":
		contentType = "application/x-www-form-urlencoded"
		newMap := make([]string, 0)
		for k, v := range param {
			vtype := reflect.TypeOf(v).String()
			if vtype == "string" {
				newMap = append(newMap, k+"="+v.(string))
			}
			if vtype == "int" {
				newMap = append(newMap, k+"="+Int2Str(v.(int)))
			}
			if vtype == "int64" {
				newMap = append(newMap, k+"="+Int642Str(v.(int64)))
			}
			if vtype == "float32" {
				newMap = append(newMap, k+"="+Float2Str(v.(float32)))
			}
			if vtype == "float64" {
				newMap = append(newMap, k+"="+Float642Str(v.(float64)))
			}
		}
		requestBody = ArrayToString(newMap, "&")
		now := genTimeNow()
		//now := "Wed, 24 Jan 2018 10:06:11 GMT"
		requestHeaders["Content-Type"] = contentType
		requestHeaders["Date"] = now
		break
	case "json":
		contentType = "application/json"
		jsonByte, _ := json.Marshal(param)
		requestBody = string(jsonByte)
		now := genTimeNow()
		//now := "Wed, 24 Jan 2018 10:06:11 GMT"
		requestHeaders["Content-Type"] = contentType
		requestHeaders["Date"] = now
		break
	default:
		break
	}
	return
}

func genTimeNow() string {
	now := time.Now()
	local, err := time.LoadLocation("GMT")
	if err != nil {
		return ""
	}
	timeNow := now.In(local).Format(time.RFC1123)
	return timeNow
}

func SetHttpTimeout(dialTimeout, dialKeepAlive, tlsHandshakeTimeout, responseHeaderTimeout, expectContinueTimeout, timeout int) HttpTimeout {
	var tt HttpTimeout

	tt.DialTimeout = dialTimeout
	tt.DialKeepAlive = dialKeepAlive
	tt.TLSHandshakeTimeout = tlsHandshakeTimeout
	tt.ResponseHeaderTimeout = responseHeaderTimeout
	tt.ExpectContinueTimeout = expectContinueTimeout
	tt.Timeout = timeout

	return tt
}

// 简单的http客户端,支持POST表单域,但不支持上传文件
func SimpleHttpClient(reqMethod string, reqUrl string, reqHeaders map[string]string, reqBody string, timeoutConf HttpTimeout) ([]byte, int, error) {
	var httpStatusCode int
	var emptyBody []byte

	req, err := http.NewRequest(reqMethod, reqUrl, strings.NewReader(reqBody))
	if err != nil {
		logs.Error("SimpleHttpClient http.NewRequest fail, reqUrl:", reqUrl)
		return emptyBody, httpStatusCode, err
	}

	for k, v := range reqHeaders {
		req.Header.Set(k, v)
	}

	client := httClientWithTimeout(timeoutConf)
	resp, err := client.Do(req)
	if err != nil {
		logs.Error("SimpleHttpClient do request fail, reqUrl:", reqUrl, ", err:", err)
		return emptyBody, httpStatusCode, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logs.Error("SimpleHttpClient read request fail, reqUrl:", reqUrl, ", err:", err)
		return emptyBody, httpStatusCode, err
	}

	return body, resp.StatusCode, err
}

// 支持post原始多文件上传,同时携带表单数据
func MultipartClient(reqUrl string, queryString map[string]string, reqHeaders map[string]string, files map[string]string, timeoutConf HttpTimeout) (originByte []byte, httpStatusCode int, err error) {
	client := httClientWithTimeout(timeoutConf)

	// 创建一个缓冲区对象,后面的要上传的body都存在这个缓冲区里
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	if len(files) <= 0 {
		logs.Error("Please select file.")
		err = fmt.Errorf("MultipartClient -> Please select file. reqUrl:%s", reqHeaders)
		return
	}
	// name: 上传表单中字段名; localABS: 待上传文件路径
	for fname, filename := range files {
		// 创建第一个需要上传的文件,filepath.Base获取文件的名称
		var fileWriter io.Writer
		fileWriter, err = bodyWriter.CreateFormFile(fname, filepath.Base(filename))
		if err != nil {
			logs.Error("The uploaded file does not exist. fname:", fname, ", filename:", filename, ", err:", err)
			return
		}
		// 打开文件
		var fd *os.File
		fd, err = os.Open(filename)
		if err != nil {
			logs.Error("Can NOT open file. fname:", fname, ", filename:", filename, ", err:", err)
			return
		}
		defer fd.Close()
		// 把第文件流写入到缓冲区里去
		_, err = io.Copy(fileWriter, fd)
		if err != nil {
			logs.Error("Can NOT copy stream. fname:", fname, ", filename:", filename, ", err:", err)
			return
		}
	}

	// 写入附加字段必须在_,_=io.Copy(fileWriter,fd)后面
	// 写入常规k,v参数
	for k, v := range queryString {
		bodyWriter.WriteField(k, v)
	}

	// 获取请求Content-Type类型,后面有用
	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	//创建一个post请求
	req, err := http.NewRequest("POST", reqUrl, nil)
	if err != nil {
		logs.Error("http.NewRequest has wrong. reqUrl:", reqUrl, ", queryString:", queryString, ", reqHeaders:", reqHeaders, ", files:", files)
		return
	}

	for k, v := range reqHeaders {
		req.Header.Set(k, v)
	}
	req.Header.Set("Content-Type", contentType)
	// 转换类型
	req.Body = ioutil.NopCloser(bodyBuf)
	// 发送数据
	resp, err := client.Do(req)
	if err != nil {
		logs.Error("clent.Do has wrong. reqUrl:", reqUrl, ", queryString:", queryString, ", reqHeaders:", reqHeaders, ", files:", files, ", err:", err)
		return
	}

	//读取请求返回的数据
	originByte, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		logs.Error("ReadAll has wrong. reqUrl:", reqUrl, ", queryString:", queryString, ", reqHeaders:", reqHeaders, ", files:", files, ", err:", err)
		return
	}
	defer resp.Body.Close()

	httpStatusCode = resp.StatusCode

	return
}

func httClientWithTimeout(timeoutConf HttpTimeout) (client *http.Client) {
	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   time.Second * time.Duration(timeoutConf.DialTimeout),
			KeepAlive: time.Second * time.Duration(timeoutConf.DialKeepAlive),
		}).Dial,
		TLSHandshakeTimeout:   time.Second * time.Duration(timeoutConf.TLSHandshakeTimeout),
		ResponseHeaderTimeout: time.Second * time.Duration(timeoutConf.ResponseHeaderTimeout),
		ExpectContinueTimeout: time.Second * time.Duration(timeoutConf.ExpectContinueTimeout),
	}
	client = &http.Client{
		Timeout:   time.Second * time.Duration(timeoutConf.Timeout),
		Transport: netTransport,
	}
	return
}

func BasicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
