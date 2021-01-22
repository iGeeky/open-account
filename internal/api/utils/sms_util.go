package utils

import (
	"open-account/pkg/baselib/log"
	"open-account/pkg/baselib/utils"
)

// SmsSend 发送短信.
func SmsSend(tel, code string) (reason string, err error) {
	log.Infof("send code [%s] to tel: %s", code, tel)
	// TODO: 支持阿里云发短信.
	return
}

// GenCode 生成验证码.
func GenCode(length int) string {
	return utils.RandomCode(length)
}
