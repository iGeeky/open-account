package middleware

import (
	// "fmt"

	"open-account/internal/api/dao"
	"open-account/internal/api/service"
	"open-account/internal/api/utils"
	"open-account/pkg/baselib/errors"
	"open-account/pkg/baselib/ginplus"
	"open-account/pkg/baselib/log"

	"time"

	"github.com/gin-gonic/gin"
)

// NeedTokenURLs 需要校验token的url字典.
var NeedTokenURLs map[string]int

func init() {
	NeedTokenURLs = make(map[string]int)
}

// TokenInfo Information contained in the token
type TokenInfo struct {
	UserType   int32
	UserID     int64
	ExpireTime int64
}

func tokenCheckInternal(token string, platform string) (tokenInfo *TokenInfo, err string) {
	userType, userID, expireTime, terr := utils.TokenDecrypt(token)
	if terr != nil {
		err = errors.ErrTokenInvalid
		return
	}

	// token 已经过期.
	if expireTime < time.Now().Unix() {
		log.Errorf("token(%s) is expired at : %d", token, expireTime)
		err = errors.ErrTokenExpired
		return
	}
	tokenErr := utils.TokenCheck(token, userID, userType, platform)
	if tokenErr != nil {
		log.Errorf("token(%s) is expired, userID: %d, expireTime: %d, err: %v", token, userID, expireTime, tokenErr)
		err = errors.ErrTokenExpired
	}

	tokenInfo = &TokenInfo{UserType: userType, UserID: userID, ExpireTime: expireTime}
	return
}

func tokenCheck(ctx *ginplus.ContextPlus, tokenCheckLevel int) bool {

	_ = ctx.Request.ParseForm()
	token := ctx.GetHeader("X-OA-Token")
	if token == "" {
		if tokenCheckLevel > service.TokenNone {
			log.Errorf("token missing..")
			ctx.JSON(401, gin.H{"ok": false, "reason": errors.ErrTokenInvalid})
			ctx.Abort()
		}
		return false
	}

	platform, _, _, _ := ctx.GetClientMetaInfo()
	// log.Infof("req_token: %s", token)
	// Check Token from account
	tokenInfo, reason := tokenCheckInternal(token, platform)
	if reason != "" {
		if tokenCheckLevel > service.TokenNone {
			log.Errorf("CheckToken(%s) failed! err: %v", token, reason)
			ctx.JSON(401, gin.H{"ok": false, "reason": reason})
			ctx.Abort()
		}
		return false
	}

	// 检查用户状态
	userDao := dao.NewUserDao()
	userInfo := userDao.MustGetByID(tokenInfo.UserID)
	if userInfo.Status == service.UserStatusDisabled {
		log.Errorf("CheckToken(%s) failed! 账号 %d:%s 已被冻结", token, userInfo.ID, userInfo.Tel)
		ctx.JsonFailWithMsg(errors.ErrUserIsLocked, "账号被冻结, 请联系客服")
		ctx.Abort()
	}

	ctx.Context.Set("userID", tokenInfo.UserID)
	ctx.Context.Set("userType", tokenInfo.UserType)
	ctx.Context.Set("token", token)
	ctx.Context.Set("userInfo", userInfo)

	return true
}

// TokenCheckFilter token verify filter.
func TokenCheckFilter(c *gin.Context) {
	if c.Request.Method == "OPTIONS" {
		c.Next()
		return
	}
	ctx := ginplus.NewContetPlus(c)
	uri := ctx.GetURI()
	// 当前url 是否需要token
	tokenCheckLevel := NeedTokenURLs[uri]

	appID, exist := c.Get("appID")
	if !exist && appID == nil {
		headers := c.Request.Header
		appID := headers.Get("X-OA-AppID")
		if appID == "" && tokenCheckLevel > service.TokenNone {
			log.Errorf("header 'X-OA-AppID not found.")
			c.JSON(401, gin.H{"ok": false, "reason": errors.ErrArgsInvalid})
			c.Abort()
			return
		}
		if appID != "" {
			ctx.Context.Set("appID", appID)
		}
	}

	tokenCheck(ctx, tokenCheckLevel)

	c.Next()
}