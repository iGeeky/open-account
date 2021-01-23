package controller

import (
	"open-account/configs"
	"open-account/internal/api/dao"
	"open-account/internal/api/service"
	"open-account/pkg/baselib/ginplus"
	"open-account/pkg/baselib/log"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	// LoginBizType 登录的bizType=login
	LoginBizType = "login"
)

// SmsLogin 短信登录.
func SmsLogin(c *gin.Context) {
	ctx := ginplus.NewContetPlus(c)
	var args SmsLoginReq
	ctx.ParseQueryJSONObject(&args)
	now := time.Now().Unix()
	ip := ctx.ClientIP()
	platform, version, channel := ctx.GetPlatformVersionChannel()

	service.CheckTelSmsCode(ctx, LoginBizType, args.Tel, args.Code)

	userType := args.UserType
	if userType == 0 {
		userType = service.UserTypeNormal
	}

	userDao := dao.NewUserDao()
	userInfo := userDao.GetByTel(args.Tel, userType)

	if userInfo == nil {
		userInfo = &dao.UserInfo{}
		userInfo.Tel = args.Tel
		userInfo.IP = ip
		userInfo.Platform = platform
		userInfo.Version = version
		userInfo.Channel = channel
		userInfo.UserType = userType
		userInfo.RegInviteCode = args.InviteCode
		userInfo.CreateTime = now
		userInfo.UpdateTime = now
		userDao.Create(userInfo, configs.Config.InviteCodeLength)
		// TODO: 添加用户注册事件.
	} else {
		userInfo.IP = ip
		userInfo.Platform = platform
		userInfo.Version = version
		userInfo.Channel = channel
		userDao.UpdateBy(userInfo, "id=?", userInfo.ID)
	}

	service.LoginInternal(ctx, userInfo)
}

// UserSmsSend 用户发送验证码.
func UserSmsSend(c *gin.Context) {
	ctx := ginplus.NewContetPlus(c)
	var args UserSmsSendReq
	ctx.ParseQueryJSONObject(&args)
	tel := args.Tel
	bizType := args.BizType
	if bizType == "" {
		bizType = "login"
	}

	code := service.GenerateCode(bizType, 5)
	ok, reason := service.SendSmsCode(bizType, tel, code)
	if !ok {
		log.Errorf("Send(bizType: %s, tel :%s) fail! err:%v, reason: %v", bizType, tel, ok, reason)
		ctx.JsonFail(reason)
		return
	}
	ctx.JsonOk(gin.H{})
}
