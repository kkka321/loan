package controllers

import (
	"crypto/md5"
	"fmt"
	"io"
	"micro-loan/common/models"
	"micro-loan/common/service"
	"micro-loan/common/tools"

	//"github.com/astaxie/beego/logs"
	"micro-loan/common/lib/gaws"

	"github.com/aws/aws-sdk-go/aws"
)

type ResourceController struct {
	BaseController
}

func (c *ResourceController) Prepare() {
	// 调用上一级的 Prepare 方法
	c.BaseController.Prepare()

	c.Data["Controller"] = "resource"
}

func (c *ResourceController) FetchImgStream() {
	rid, _ := tools.Str2Int64(c.Ctx.Input.Param(":rid"))
	//logs.Debug("rid:", rid)

	resource, err := service.OneResource(rid)
	if err != nil {
		c.Ctx.Output.Status = 404
		return
	}

	etag := c.Ctx.Request.Header.Get("If-None-Match")
	//logs.Debug("If-None-Match:", etag)
	if etag == resource.ContentMd5 {
		c.Ctx.Output.Status = 304
		return
	}

	var b []byte
	w := aws.NewWriteAtBuffer(b)
	_, err = gaws.AwsDownload2Stream(resource.HashName, w)
	if err != nil {
		c.Ctx.Output.Status = 404
		return
	}

	mime := "image/png"
	if len(resource.Mime) > 4 {
		mime = resource.Mime
	}
	c.Ctx.Output.Header("Content-Type", mime)
	c.Ctx.Output.Header("Etag", resource.ContentMd5)
	c.Ctx.Output.Body(w.Bytes())
}

func (c *ResourceController) FetchAudioStream() {
	rid, _ := tools.Str2Int64(c.Ctx.Input.Param(":rid"))

	callRecord, err := models.GetSipCallRecordById(rid)
	if err != nil {
		c.Ctx.Output.Status = 404
		return
	}

	h := md5.New()
	io.WriteString(h, callRecord.AudioRecordName)
	nameMd5 := fmt.Sprintf("%x", h.Sum(nil))

	etag := c.Ctx.Request.Header.Get("If-None-Match")
	//logs.Debug("If-None-Match:", etag)
	if etag == nameMd5 {
		c.Ctx.Output.Status = 304
		return
	}

	var b []byte
	w := aws.NewWriteAtBuffer(b)
	_, err = gaws.AwsDownload2Stream(tools.CreateVoipFileName(callRecord.AudioRecordName), w)
	if err != nil {
		c.Ctx.Output.Status = 404
		return
	}

	mime := "audio/mpeg"

	c.Ctx.Output.Header("Content-Type", mime)
	c.Ctx.Output.Header("Etag", nameMd5)
	c.Ctx.Output.Body(w.Bytes())
}
