package controller

import (
	"open-account/configs"
	"open-account/internal/api/service"
	"open-account/internal/api/utils"
	"open-account/pkg/baselib/cache"
	"open-account/pkg/baselib/errors"
	"open-account/pkg/baselib/ginplus"
	"open-account/pkg/baselib/log"

	"github.com/gin-gonic/gin"
)

// SmsCheck 短信检查, 后台接口,提供给其它应用内部使用的.
func SmsCheck(c *gin.Context) {
	ctx := ginplus.NewContetPlus(c)
	var args SMSCheckReq
	ctx.ParseQueryJSONObject(&args)

	ok, reason := service.CheckSmsCode(args.BizType, args.Tel, args.Code)
	if !ok {
		log.Errorf("check(bizType: %s, tel :%s) fail! err:%v, reason: %v", args.BizType, args.Tel, ok, reason)
		ctx.JsonFail(reason)
		return
	}
	ctx.JsonOk(gin.H{})
}

// SmsGetCode 查询短信验证码(用于内部测试使用, 不要暴露接口在外部环境中.)
func SmsGetCode(c *gin.Context) {
	ctx := ginplus.NewContetPlus(c)
	bizType := ctx.MustGet("bizType")
	tel := ctx.MustGet("tel")
	key := ctx.MustGet("key")

	var err error

	superKey := configs.Config.SuperKeyForTest
	if key != superKey {
		reason := "unauthorized"
		ctx.JsonFail(reason)
		return
	}

	// tel := cc + "-" + tel
	log.Infof("----------- Get Sms Code, bizType: %s, tel: %s", bizType, tel)

	var code string
	code, err = utils.GetCode(bizType, tel)
	if err != nil && err != cache.ErrNotExist {
		log.Errorf("GetCode(%s, %s) failed! err: %v", bizType, code, err)
		ctx.JsonFail(errors.ErrServerError)
		return
	}
	var reason string
	if err == cache.ErrNotExist {
		reason = "not found"
		ctx.JsonFail(reason)
	} else {
		ctx.JsonOk(gin.H{"code": code})
	}
}
