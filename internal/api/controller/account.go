package controller

import (
	"open-account/configs"
	"open-account/internal/api/dao"
	"open-account/internal/api/service"
	"open-account/pkg/baselib/errors"
	"open-account/pkg/baselib/ginplus"
	"open-account/pkg/baselib/log"
	"open-account/pkg/baselib/utils"
	"strings"
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
	platform, version, channel, deviceID := ctx.GetClientMetaInfo()

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
		userInfo.DeviceID = deviceID
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
		bizType = LoginBizType
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

// UserRegister 用户手机号注册.
func UserRegister(c *gin.Context) {
	ctx := ginplus.NewContetPlus(c)
	var args UserRegisterReq
	ctx.ParseQueryJSONObject(&args)
	if args.UserType == 0 {
		args.UserType = service.UserTypeNormal
	}
	now := utils.Now()
	ip := ctx.ClientIP()
	platform, version, channel, deviceID := ctx.GetClientMetaInfo()
	service.CheckTelSmsCode(ctx, LoginBizType, args.Tel, args.Code)

	args.Tel = strings.Replace(args.Tel, " ", "", -1)
	args.Tel = strings.Replace(args.Tel, "\t", "", -1)
	args.Nickname = strings.TrimSpace(args.Nickname)

	userDao := dao.NewUserDao()
	existUserInfo := userDao.GetByTel(args.Tel, args.UserType)
	if existUserInfo != nil {
		ctx.JsonFail(errors.ErrTelRegisted)
		return
	}

	if args.Username != "" {
		existUserInfo = userDao.GetByUsername(args.Username, args.UserType)
		if existUserInfo != nil {
			ctx.JsonFail(errors.ErrUsernameRegisted)
			return
		}
	}

	encodePassword := utils.MustPwdEncode(args.Password)

	userInfo := &dao.UserInfo{}
	userInfo.Tel = args.Tel
	userInfo.Password = encodePassword
	userInfo.Username = args.Username
	userInfo.Nickname = args.Nickname
	userInfo.UserType = args.UserType
	userInfo.IP = ip
	userInfo.Platform = platform
	userInfo.Version = version
	userInfo.Channel = channel
	userInfo.DeviceID = deviceID
	userInfo.RegInviteCode = args.InviteCode
	userInfo.CreateTime = now
	userInfo.UpdateTime = now
	if len(args.Profile) > 0 {
		err := userInfo.Profile.Scan([]byte(args.Profile))
		if err != nil {
			log.Errorf("Profile.Scan(%s) failed! err: %v", string(args.Profile), err)
			errors.Assert(err)
		}
	}

	userDao.Create(userInfo, configs.Config.InviteCodeLength)
	// TODO: 添加用户注册事件.

	service.LoginInternal(ctx, userInfo)
}

// UserCheckTelExist 检查手机号是否存在
func UserCheckTelExist(c *gin.Context) {
	ctx := ginplus.NewContetPlus(c)
	tel := ctx.MustGet("tel")
	userType := ctx.GetInt32("userType", service.UserTypeNormal)
	dao := dao.NewUserDao()
	userInfo := dao.GetByTel(tel, userType)
	exist := (userInfo != nil)
	ctx.JsonOk(gin.H{"tel": tel, "userType": userType, "exist": exist})
}

// UserCheckUsernameExist 检查手机号是否存在
func UserCheckUsernameExist(c *gin.Context) {
	ctx := ginplus.NewContetPlus(c)
	username := ctx.MustGet("username")
	userType := ctx.GetInt32("userType", service.UserTypeNormal)
	dao := dao.NewUserDao()
	userInfo := dao.GetByUsername(username, userType)
	exist := (userInfo != nil)
	ctx.JsonOk(gin.H{"username": username, "userType": userType, "exist": exist})
}
