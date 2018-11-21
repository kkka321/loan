package controllers

import (
	"github.com/astaxie/beego/logs"
	"micro-loan/common/pkg/short_url"
)

type ShortUrlController struct {
	ApiBaseController
}

func (c *ShortUrlController) Prepare() {
	// 调用上一级的 Prepare 方
	//c.ApiBaseController.Prepare()

	// 统一将 ip 加到 RequestJSON 中
	//c.RequestJSON["ip"] = c.Ctx.Input.IP()
	//c.RequestJSON["related_id"] = int64(0)
}

func (c *ShortUrlController) Access() {
	shortUrl := c.Ctx.Input.Param(":u")
	if shortUrl == "" {
		logs.Warning("[Access] shortUrl Empty")

		c.Ctx.Output.Header("Content-Type", "text/html")
		c.Ctx.Output.Body([]byte("url not found"))
		c.Ctx.Output.Status = 404
		return
	}

	urlInfo, err := short_url.GetShortUrl(shortUrl)
	if err != nil {
		logs.Warning("[Access] GetShortUrl record not found url:%s, err:%v", shortUrl, err)

		c.Ctx.Output.Header("Content-Type", "text/html")
		c.Ctx.Output.Body([]byte("url not found"))
		c.Ctx.Output.Status = 404
		return
	}

	logs.Warning("[Access] Redirect shorturl:%s, url:%s", shortUrl, urlInfo.Url)
	c.Redirect(urlInfo.Url, 302)
}
