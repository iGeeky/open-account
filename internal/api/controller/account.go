package controller

import (
	"github.com/iGeeky/open-account/configs"
	"github.com/iGeeky/open-account/internal/api/dao"
	"github.com/iGeeky/open-account/internal/api/service"
	apiutils "github.com/iGeeky/open-account/internal/api/utils"
	"github.com/iGeeky/open-account/pkg/baselib/errors"
	"github.com/iGeeky/open-account/pkg/baselib/ginplus"
	"github.com/iGeeky/open-account/pkg/baselib/log"
	"github.com/iGeeky/open-account/pkg/baselib/utils"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sevlyar/retag"
)

const (
	// LoginBizType 登录的bizType=login
	LoginBizType = "login"
	// ResetPwdBizType = "resetPwd"
	ResetPwdBizType = "resetPwd"
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
		userInfo.Status = service.UserStatusOK
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
	service.CheckCaptcha(ctx, args.CaptchaID, args.CaptchaValue, "tel="+tel)

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
	userInfo.Status = service.UserStatusOK
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

// UserLogin 用户密码登录
func UserLogin(c *gin.Context) {
	ctx := ginplus.NewContetPlus(c)
	var args UserLoginReq
	ctx.ParseQueryJSONObject(&args)
	if args.UserType == 0 {
		args.UserType = service.UserTypeNormal
	}
	service.CheckCaptcha(ctx, args.CaptchaID, args.CaptchaValue, "tel="+args.Tel)

	userDao := dao.NewUserDao()
	var userInfo *dao.UserInfo
	if args.Tel != "" {
		userInfo = userDao.GetByTel(args.Tel, args.UserType)
		if userInfo == nil {
			log.Infof("user {tel: %s, userType: %d} login failed! tel not exist!", args.Tel, args.UserType)
			ctx.JsonFail(errors.ErrTelNotExist)
			return
		}
	} else if args.Username != "" {
		userInfo = userDao.GetByUsername(args.Username, args.UserType)
		if userInfo == nil {
			log.Infof("user {username: %s, userType: %d} login failed! username not exist!", args.Username, args.UserType)
			ctx.JsonFail(errors.ErrUsernameNotExist)
			return
		}
	} else {
		log.Warnf("user %+v login failed! ", &args)
		ctx.JsonFail(errors.ErrArgsInvalid)
		return
	}

	// 比较密码.
	err := utils.PwdCompare(userInfo.Password, args.Password)
	if err != nil {
		ctx.JsonFail(errors.ErrPasswordErr)
		return
	}

	service.LoginInternal(ctx, userInfo)
}

// UserLogout 用户登出
func UserLogout(c *gin.Context) {
	ctx := ginplus.NewContetPlus(c)
	userID, _ := ctx.MustGetUserID()
	token := ctx.GetCustomHeader("token")
	_ = apiutils.TokenDelete(userID, token)
	ctx.JsonOk(gin.H{})
}

// UserGetInfo 获取用户基本信息.
func UserGetInfo(c *gin.Context) {
	ctx := ginplus.NewContetPlus(c)
	userID, _ := ctx.MustGetUserID()
	dao := dao.NewUserDao()
	userInfo := retag.Convert(dao.MustGetByID(userID), retag.NewView("json", "detail"))
	ctx.JsonOk(gin.H{"userInfo": userInfo})
}

// UserSetInfo 设置用户基本信息
func UserSetInfo(c *gin.Context) {
	ctx := ginplus.NewContetPlus(c)
	userID, _ := ctx.MustGetUserID()
	var args UserSetInfoReq
	ctx.ParseQueryJSONObject(&args)

	dao := dao.NewUserDao()
	userInfo := dao.MustGetByID(userID)
	if args.Avatar != "" {
		userInfo.Avatar = args.Avatar
	}
	if args.Nickname != "" {
		userInfo.Nickname = args.Nickname
	}
	if args.Sex != 0 {
		userInfo.Sex = args.Sex
	}
	if args.Birthday != "" {
		userInfo.Birthday = args.Birthday
	}
	dao.UpdateBy(userInfo, "id = ?", userID)

	ctx.JsonOk(gin.H{"userInfo": retag.Convert(userInfo, retag.NewView("json", "detail"))})
}

// UserSetName 设置用户名
func UserSetName(c *gin.Context) {
	ctx := ginplus.NewContetPlus(c)
	userID, _ := ctx.MustGetUserID()
	var args UserSetNameReq
	ctx.ParseQueryJSONObject(&args)

	dao := dao.NewUserDao()
	userInfo := dao.MustGetByID(userID)
	if userInfo.Username != "" {
		log.Infof("user {tel: %s, userType: %d} already has a username and cannot be set again!", userInfo.Tel, userInfo.UserType)
		ctx.JsonFail(errors.ErrUserHaveUsername)
		return
	}

	dao.SetUsername(userID, args.Username)

	ctx.JsonOk(gin.H{})
}

// UserChangePassword 用户修改密码
func UserChangePassword(c *gin.Context) {
	ctx := ginplus.NewContetPlus(c)
	userID, _ := ctx.MustGetUserID()
	var args UserChangePasswordReq
	ctx.ParseQueryJSONObject(&args)

	userDao := dao.NewUserDao()
	userInfo := userDao.MustGetByID(userID)

	// verify old password
	err := utils.PwdCompare(userInfo.Password, args.OldPassword)
	if err != nil {
		ctx.JsonFail(errors.ErrPasswordErr)
		return
	}

	encodePassword := utils.MustPwdEncode(args.Password)
	userDao.SetPassword(userInfo.ID, encodePassword)

	ctx.JsonOk(gin.H{})
}

// UserResetPassword 用户重置密码.
func UserResetPassword(c *gin.Context) {
	ctx := ginplus.NewContetPlus(c)
	var args UserResetPasswordReq
	ctx.ParseQueryJSONObject(&args)
	service.CheckTelSmsCode(ctx, ResetPwdBizType, args.Tel, args.Code)

	if args.UserType == 0 {
		args.UserType = service.UserTypeNormal
	}

	userDao := dao.NewUserDao()
	// check tel, username exist
	userInfo := userDao.GetByTel(args.Tel, args.UserType)
	if userInfo == nil {
		log.Infof("Tel [%s, type: %d] is not exist!", args.Tel, args.UserType)
		ctx.JsonFail(errors.ErrTelNotExist)
		return
	}

	encodePassword := utils.MustPwdEncode(args.Password)
	userDao.SetPassword(userInfo.ID, encodePassword)

	service.LoginInternal(ctx, userInfo)
}

// InviteCodeSettable 检查是否可以设置邀请码.
func InviteCodeSettable(c *gin.Context) {
	ctx := ginplus.NewContetPlus(c)
	var settable = true
	userID, _ := ctx.MustGetUserID()
	userDao := dao.NewUserDao()
	userInfo := userDao.MustGetByID(userID)
	if (utils.Now() - userInfo.CreateTime) > int64(configs.Config.InviteCodeSettingPeriod.Seconds()) { //注册时间超过设置的限制.
		settable = false
	} else if userInfo.RegInviteCode != "" { // 已经设置
		settable = false
	}

	ctx.JsonOk(gin.H{"settable": settable})
}

// SetInviteCode 设置注册邀请码.
func SetInviteCode(c *gin.Context) {
	ctx := ginplus.NewContetPlus(c)
	userID, _ := ctx.MustGetUserID()
	var args SetInviteCodeReq
	ctx.ParseQueryJSONObject(&args)
	userDao := dao.NewUserDao()
	userInfo := userDao.MustGetByID(userID)
	if userInfo.RegInviteCode != "" { //注册邀请码 已经设置过.
		ctx.JsonOk(gin.H{})
		return
	}
	args.InviteCode = strings.ToLower(args.InviteCode)

	inviteUserInfo := userDao.GetByInviteCode(args.InviteCode)
	if inviteUserInfo == nil {
		log.Errorf("user %s(%d) input a invite invite code: %s", userInfo.Username, userInfo.ID, args.InviteCode)
		ctx.JsonFail(errors.ErrInviteCodeInvalid)
		return
	}

	// 不能使用自己的注册码.
	if inviteUserInfo.ID == userInfo.ID {
		log.Infof("User %s(%d) used his own inviteCode", userInfo.Username, userInfo.ID)
		ctx.JsonFailWithMsg(errors.ErrInviteCodeInvalid, "不能使用自己的邀请码")
		return
	}

	userDao.SetRegInviteCode(userID, args.InviteCode)

	ctx.JsonOk(gin.H{})
}
