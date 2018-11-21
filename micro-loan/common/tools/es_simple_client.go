package tools

import (
	"fmt"

	"encoding/json"

	"github.com/astaxie/beego/logs"
)

var esHttpRequestHeader = map[string]string{
	"Content-Type": "application/json",
}

// EsDeleteIndex 删除某一索引下全部数据
func EsDeleteIndex(esHost, esIndex string) (result bool, err error) {
	esApi := fmt.Sprintf(`%s/%s`, esHost, esIndex)
	httpBody, httpStatusCode, err := SimpleHttpClient("DELETE", esApi, esHttpRequestHeader, "", DefaultHttpTimeout())

	logs.Debug("[EsDeleteIndex] esApi:", esApi, ", httpStatusCode", httpStatusCode, ", httpBody:", string(httpBody))

	if httpStatusCode == 200 {
		result = true
	} else if httpStatusCode == 404 {
		err = fmt.Errorf("no such index [%s]", esIndex)
	}

	return
}

// EsCreate 几 es 中写数据
func EsCreate(esHost, esIndex, esType, dataJSON string) (result bool, id string, err error) {
	esApi := fmt.Sprintf(`%s/%s/%s`, esHost, esIndex, esType)
	httpBody, httpStatusCode, err := SimpleHttpClient("POST", esApi, esHttpRequestHeader, dataJSON, DefaultHttpTimeout())

	logs.Debug("[EsCreate] esApi:", esApi, ", httpStatusCode", httpStatusCode, ", httpBody:", string(httpBody))

	if httpStatusCode == 201 {
		result = true

		var response = map[string]interface{}{}
		errJ := json.Unmarshal(httpBody, &response)
		if errJ != nil {
			err = errJ
			return
		}
		if v, ok := response["_id"]; ok {
			id = v.(string)
		}
	}

	return
}

// EsSetMappings 手工设置 mappings
func EsSetMappings(esHost, esIndex, mappings string) (result bool, err error) {
	esApi := fmt.Sprintf(`%s/%s`, esHost, esIndex)
	httpBody, httpStatusCode, err := SimpleHttpClient("PUT", esApi, esHttpRequestHeader, mappings, DefaultHttpTimeout())

	logs.Debug("[EsSetMappings] esApi:", esApi, ", httpStatusCode", httpStatusCode, ", httpBody:", string(httpBody))

	if httpStatusCode == 400 {
		err = fmt.Errorf("index [%s] already exists", esIndex)
	} else if httpStatusCode == 200 {
		result = true
	} else {
		err = fmt.Errorf("unknow err, htt status code: %d, err: %v", httpStatusCode, err)
	}

	return
}

// EsIndexIsExist 查看索引的mappings是否存在,以些判断索引是否存在
func EsIndexIsExist(esHost, esIndex string) (yes bool, err error) {
	esApi := fmt.Sprintf(`%s/%s/_mappings`, esHost, esIndex)
	httpBody, httpStatusCode, err := SimpleHttpClient(HttpMethodGet, esApi, map[string]string{}, "", DefaultHttpTimeout())

	logs.Debug("[EsIndexIsExist] esApi:", esApi, ", httpBody:", string(httpBody))

	if err != nil {
		return
	} else if httpStatusCode == 200 {
		yes = true
	} else if httpStatusCode == 404 {
		err = fmt.Errorf("no such index [%s]", esIndex)
	}

	return
}
