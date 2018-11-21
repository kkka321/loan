package main

import (
	"micro-loan/common/tools"

	"github.com/astaxie/beego/logs"
)

var (
	esHost  = "http://52.221.78.48:9200"
	esIndex = "nginx"
	esType  = "api-log"
)

func main() {
	//testEsIndexIsExist()
	//testEsDeleteIndex()
	testEsCreate()
}

func testEsCreate() {
	dataJSON := `{"@timestamp":"2018-05-26T07:21:02+00:00","@version":1,"server_addr":"10.106.2.58","remote_addr":"61.5.33.236","http_x_forwarded_for":"61.5.33.236","body_bytes_sent":25,"request_time":0.001,"host":"api.rupiahcepatweb.com","request_method":"POST","request_uri":"/api/v1/account/info","server_protocol":"HTTP/1.1","http_referer":"-","http_user_agent":"okhttp/3.9.1 (Android 5.1) 10048/v/3","upstream_response_time":"0.001","status":200}`
	result, id, err := tools.EsCreate(esHost, esIndex, esType, dataJSON)
	logs.Debug("result:", result, ", id: ", id, ", err:", err)

}

func testEsDeleteIndex() {
	result, err := tools.EsDeleteIndex(esHost, esIndex)
	logs.Debug("result:", result, ", err:", err)
}

func testEsIndexIsExist() {
	yes, err := tools.EsIndexIsExist(esHost, esIndex)
	logs.Debug("yes:", yes, ", err:", err)

	if !yes {
		testEsSetMappings()
	}
}

func testEsSetMappings() {
	mappings := `{
    "mappings" : {
      "api-log" : {
        "properties" : {
          "@timestamp" : {
            "type" : "date"
          },
          "@version" : {
            "type" : "integer"
          },
          "body_bytes_sent" : {
            "type" : "long"
          },
          "remote_addr" : {
            "type" : "ip"
          },
          "request_time" : {
            "type" : "float"
          },
          "server_addr" : {
            "type" : "ip"
          },
          "status" : {
            "type" : "integer"
          },
          "upstream_response_time" : {
            "type" : "float"
          }
        }
      }
    }
}`

	result, err := tools.EsSetMappings(esHost, esIndex, mappings)
	logs.Debug("result:", result, ", err:", err)
}
