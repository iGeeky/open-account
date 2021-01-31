package controller

import (
	"open-account/configs"
	"open-account/internal/api/dao"
	"open-account/internal/api/service"
	"open-account/pkg/baselib/errors"
	"open-account/pkg/baselib/ginplus"
	"open-account/pkg/baselib/log"
	"open-account/pkg/baselib/utils"

	"github.com/gin-gonic/gin"
)

// ManagerUserPasswordReset 后台重置密码
func ManagerUserPasswordReset(c *gin.Context) {
	ctx := ginplus.NewContetPlus(c)
	var args ManagerUserResetPwdReq
	ctx.ParseQueryJSONObject(&args)
	if args.UserType == 0 {
		args.UserType = service.UserTypeNormal
	}
	userDao := dao.NewUserDao()
	userInfo := userDao.GetByTel(args.Tel, args.UserType)
	if userInfo == nil {
		ctx.JsonFailWithMsg(errors.ErrTelNotExist, "手机号不存在")
		return
	}

	cliPassword := args.Password
	encodePassword := utils.MustPwdEncode(cliPassword)
	log.Infof("origin pwd: %s, encodePassword: %s", args.Password, encodePassword)
	userDao.SetPassword(userInfo.ID, encodePassword)

	ctx.JsonOk(gin.H{})
}

// ManagerUserDeRegister 后台删除用户.
func ManagerUserDeRegister(c *gin.Context) {
	ctx := ginplus.NewContetPlus(c)
	var args ManagerUserDeRegisterReq
	ctx.ParseQueryJSONObject(&args)
	if configs.Config.Debug {
		userDao := dao.NewUserDao()
		var userInfo *dao.UserInfo
		if args.ID > 0 {
			userInfo = userDao.MustGetByID(args.ID)
		} else {
			ctx.JsonOk(gin.H{})
		}
		userDao.DeleteByID(userInfo.ID)

		log.Infof("user [%v] deleted!", userInfo.ID)
	}
	ctx.JsonOk(gin.H{})
}

// ManagerUserSetStatus 后台设置用户状态
func ManagerUserSetStatus(c *gin.Context) {
	ctx := ginplus.NewContetPlus(c)
	var args ManagerUserSetStatusReq
	ctx.ParseQueryJSONObject(&args)

	dao := dao.NewUserDao()
	dao.MustGetByID(args.ID)

	dao.SetStatus(args.ID, args.Status)

	ctx.JsonOk(gin.H{})
}
