package controller

import (
	"fmt"
	"open-account/pkg/baselib/ginplus"
	"open-account/pkg/baselib/utils"

	"github.com/dchest/captcha"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
)

// APIPing ping测试接口
func APIPing(c *gin.Context) {
	ctx := ginplus.NewContetPlus(c)
	data := gin.H{"buildTime": utils.BuildTime, "GitBranch": utils.GitBranch, "GitCommit": utils.GitCommit, "now": utils.Datetime()}
	ctx.JsonOk(data)
}

// CaptchaInfo 获取captcha信息
func CaptchaInfo(c *gin.Context) {
	ctx := ginplus.NewContetPlus(c)
	captchaID := captcha.New()
	captchaImg := fmt.Sprintf("/v1/account/captcha/%s.png", captchaID)
	ctx.JsonOk(gin.H{"captchaID": captchaID, "captchaImg": captchaImg})
}

// CaptchaFile Captcha文件下载地址.
func CaptchaFile(c *gin.Context) {
	name := c.Param("name")
	var captchatype string
	var captchaID string
	if len(name) <= 4 {
		c.HTML(400, "BAD REQUEST", nil)
		return
	}

	captchatype = name[len(name)-3:]
	captchaID = name[:len(name)-4]
	if captchatype == "png" {
		err := captcha.WriteImage(c.Writer, captchaID, 360, 120)
		if err != nil {
			glog.Errorf("captcha.WriteImage failed! err: %v", err)
			c.HTML(500, "SERVER ERROR", nil)
			return
		}
	} else if captchatype == "wav" {
		var AudioLang = "zh"
		err := captcha.WriteAudio(c.Writer, captchaID, AudioLang)
		if err != nil {
			glog.Errorf("captcha.WriteAudio failed! err: %v", err)
			c.HTML(500, "SERVER ERROR", nil)
			return
		}
	} else {
		c.HTML(400, "BAD REQUEST", nil)
		return
	}
}
