package controllers

import (
	"fmt"
	"micro-loan/common/cerror"
	"micro-loan/common/tools"

	"github.com/astaxie/beego/logs"
)

type DotController struct {
	DotBaseController
}

func (c *DotController) Prepare() {
	// 调用上一级的 Prepare 方
	c.DotBaseController.Prepare()

}

func (c *DotController) Dot1() {
	var ver string
	var uiver string
	var parae string
	if v, ok := c.RequestJSON["ver"]; ok {
		if v != nil {
			ver = v.(string)
		}
	}
	if v, ok := c.RequestJSON["uiver"]; ok {
		if v != nil {
			uiver = v.(string)
		}
	}
	if v, ok := c.RequestJSON["parae"]; ok {
		if v != nil {
			parae = v.(string)
		}
	}
	//1. 验证参数不为空
	if ver == "" || uiver == "" || parae == "" {
		code := cerror.LostRequiredParameters
		c.Data["json"] = cerror.BuildApiResponse(code, "")
		c.ServeJSON()
		return
	}
	//2.转发请求到打点服务
	url := fmt.Sprintf(Dot1, ver, uiver, parae)
	reqHeaders := map[string]string{
		"Connection":   "keep-alive",
		"Content-Type": "application/x-www-form-urlencoded",
		"User-Agent":   "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.132 Safari/537.36",
	}
	logs.Debug("[Dot1] URL is: ", url)
	httpBody, httpStatusCode, err := tools.SimpleHttpClient("GET", url, reqHeaders, "", tools.DefaultHttpTimeout())
	logs.Debug("[Dot1] Result httpBody: %s, httpStatusCode: %d, err: %v\n", httpBody, httpStatusCode, err)
	//3. 返回结果给客户端
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, "")
	c.ServeJSON()

}
func (c *DotController) Dot2() {
	var pkg string
	var ver string
	var uiver string
	var content string
	if v, ok := c.RequestJSON["pkg"]; ok {
		if v != nil {
			pkg = v.(string)
		}
	}
	if v, ok := c.RequestJSON["ver"]; ok {
		if v != nil {
			ver = v.(string)
		}
	}
	if v, ok := c.RequestJSON["uiver"]; ok {
		if v != nil {
			uiver = v.(string)
		}
	}
	if v, ok := c.RequestJSON["content"]; ok {
		if v != nil {
			content = v.(string)
		}
	}
	logs.Debug("[Dot2] 参数content:", content)
	//1. 验证参数不为空
	if ver == "" || uiver == "" || pkg == "" || content == "" {
		code := cerror.LostRequiredParameters
		c.Data["json"] = cerror.BuildApiResponse(code, "")
		c.ServeJSON()
		return
	}
	//2.转发请求到打点服务
	url := fmt.Sprintf(Dot2, pkg, ver, uiver)
	reqHeaders := map[string]string{
		"Connection":   "keep-alive",
		"Content-Type": "application/x-www-form-urlencoded",
		"User-Agent":   "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.132 Safari/537.36",
	}
	logs.Debug("[Dot2] URL is: ", url)

	httpBody, httpStatusCode, err := tools.SimpleHttpClient("POST", url, reqHeaders, content, tools.DefaultHttpTimeout())
	logs.Debug("[Dot2] Result httpBody: %s, httpStatusCode: %d, err: %v\n", httpBody, httpStatusCode, err)
	//3. 返回结果给客户端
	c.Data["json"] = cerror.BuildApiResponse(cerror.CodeSuccess, "")
	c.ServeJSON()
}
