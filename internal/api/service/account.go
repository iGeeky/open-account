package service

import (
	"open-account/configs"
	"open-account/internal/api/dao"
	"open-account/internal/api/utils"
	"open-account/pkg/baselib/errors"
	"open-account/pkg/baselib/ginplus"
	"open-account/pkg/baselib/log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sevlyar/retag"
)

func tokenCreate(userInfo *dao.UserInfo, platform string) (token string, err error) {
	userID, userType := userInfo.ID, userInfo.UserType
	log.Infof("login >>> userID: %d, userType: %d uid: %s", userID, userType, userInfo.UID)

	tokenTimeout := configs.Config.TokenRedis.Exptime
	expires := time.Now().Unix() + int64(tokenTimeout.Seconds())
	token = utils.TokenEncrypt(userType, userID, expires)
	err = utils.TokenSave(token, userID, userType, platform, tokenTimeout)
	return
}

func CheckTelSmsCode(ctx *ginplus.ContextPlus, bizType, tel, code string) (ok bool) {
	ok = true
	ignoreCodeVerify := configs.Config.Debug && code == configs.Config.SuperCodeForTest
	testCode, isTestAccount := configs.Config.TestAccounts[tel]
	if isTestAccount && code == testCode {
		log.Infof("Test account {tel: %s, code: %s}", tel, code)
		ignoreCodeVerify = true
	}

	//校验验证码
	if ignoreCodeVerify {
		ip := ctx.ClientIP()
		log.Infof("bizType: %s, tel: %s, code: %s ip: %s ignore code verify", bizType, tel, code, ip)
	} else {
		ok, reason := CheckSmsCode(bizType, tel, code)
		if !ok {
			panic(errors.NewError(reason, "验证码错误"))
		}
	}
	return
}

func LoginInternal(ctx *ginplus.ContextPlus, userInfo *dao.UserInfo) {
	if userInfo.Status == UserStatusDisabled {
		ctx.JsonFailWithMsg(errors.ErrUserIsLocked, "账号被冻结, 请联系客服")
		return
	}

	platform, _, _ := ctx.GetPlatformVersionChannel()
	token, err := tokenCreate(userInfo, platform)
	if err != nil {
		ctx.JsonFailWithMsg(errors.ErrServerError, "服务器出问题了, 请稍后再试")
		return
	}
	log.Infof("user [%s] login success!", userInfo.Tel)

	data := gin.H{"token": token, "userInfo": retag.Convert(userInfo, retag.NewView("json", "detail"))}
	ctx.JsonOk(data)
}
