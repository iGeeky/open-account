package service

import (
	"github.com/iGeeky/open-account/configs"
	"github.com/iGeeky/open-account/internal/api/utils"
	"github.com/iGeeky/open-account/pkg/baselib/cache"
	"github.com/iGeeky/open-account/pkg/baselib/errors"
	"github.com/iGeeky/open-account/pkg/baselib/log"
)

// GenerateCode 生成验证码.
func GenerateCode(bizType string, length int) string {
	return utils.GenCode(length)
}

// SendSmsCode 发送短信,并保存到redis.
func SendSmsCode(bizType, tel, code string) (ok bool, reason string) {
	// 如果没有传入过期时间, 默认为5分钟.
	expires := configs.Config.SmsRedis.Exptime
	log.Infof("----------- Send Sms Code, bizType: %s, tel: %s, Code: %s", bizType, tel, code)
	// 使用运营商短信服务发送短信
	reason, err := utils.SmsSend(tel, code)
	if err != nil {
		log.Errorf("SmsSend(%s, %s) failed! err: %v, reason: %s", tel, code, err, reason)
		return false, errors.ErrServerError
	}

	if reason != "" {
		return false, reason
	}

	// 保存验证码
	err = utils.SaveCode(bizType, tel, code, expires)
	if err != nil {
		log.Errorf("saveTel(%s, %s,%s,%d) failed! err: %v", bizType, tel, code, expires, err)
		return false, errors.ErrServerError
	}
	return true, ""
}

// CheckSmsCode 检查短信.
func CheckSmsCode(bizType, tel string, reqCode string) (ok bool, reason string) {
	log.Infof("----------- Check Sms Code, bizType: %s, tel: %s, Code: %s", bizType, tel, reqCode)
	code, err := utils.GetCode(bizType, tel)
	if err != nil {
		if err == cache.ErrNotExist {
			log.Errorf("GetCode(%s, %s) failed! err: code not exist!", bizType, tel)
			return false, errors.ErrCodeInvalid
		} else {
			log.Errorf("GetCode(%s, %s) failed! err: %v", bizType, tel, err)
			return false, errors.ErrServerError
		}
	}

	if code != reqCode {
		log.Errorf("----------- Check Sms Code, bizType: %s, tel: %s, Code: %s failed! server code: %s", bizType, tel, reqCode, code)
		return false, errors.ErrCodeInvalid
	}
	return true, ""
}
