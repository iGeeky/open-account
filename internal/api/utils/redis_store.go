package utils

import (
	"fmt"
	"open-account/configs"
	"open-account/pkg/baselib/cache"
	"open-account/pkg/baselib/log"
	"strconv"
	"time"
)

var (
	tokenRedis *cache.RedisCache
	smsRedis   *cache.RedisCache
)

// InitRedisStore 初始化Redis存储.
func InitRedisStore(Config *configs.ServerConfig) (err error) {
	tokenRedis, err = cache.NewRedisCache(Config.TokenRedis)
	if err != nil {
		log.Fatalf("NewRedisCache(%s) failed! err: %v", Config.TokenRedis.DebugStr(), err)
		return
	}
	smsRedis, err = cache.NewRedisCache(Config.SmsRedis)
	if err != nil {
		log.Fatalf("NewRedisCache(%s) failed! err: %v", Config.SmsRedis.DebugStr(), err)
		return
	}

	return
}

// TokenSave 保存Token
func TokenSave(token string, userID int64, userType int32, platform string, tokenTimeout time.Duration) (err error) {
	key := fmt.Sprintf("tk:%s", token)
	err = tokenRedis.Set(key, userID, tokenTimeout)
	log.Infof("token.save('%s', '%d', timeout: %v). err: %v", key, userID, tokenTimeout, err)

	return
}

// TokenCheck 检查Token是否存在,
func TokenCheck(token string, userID int64, userType int32, platform string) (err error) {
	key := fmt.Sprintf("tk:%s", token)
	var strUserID string
	strUserID, err = tokenRedis.Get(key)

	if err == cache.ErrNotExist || strUserID == "" {
		err = fmt.Errorf("TOKEN_NOT_EXIST")
	} else {
		userIDInDB, _ := strconv.ParseInt(strUserID, 10, 64)
		if userIDInDB != userID {
			err = fmt.Errorf("TOKEN_NOT_EXIST")
		} else {
			err = nil
		}
	}

	return
}

// TokenDelete 删除Token
func TokenDelete(userID int64, token string) (err error) {
	key := fmt.Sprintf("tk:%d", userID)
	_, err = tokenRedis.Del(key)
	if err != nil {
		return
	}
	return
}

// SaveCode 保存验证码
func SaveCode(bizType, tel, code string, expires time.Duration) (err error) {
	key := "cd:" + bizType + ":" + tel
	err = smsRedis.Set(key, code, expires)
	return
}

// GetCode 查询验证码.
func GetCode(bizType, tel string) (code string, err error) {
	key := "cd:" + bizType + ":" + tel
	code, err = smsRedis.Get(key)
	return
}
