package utils

import (
	"github.com/iGeeky/open-account/pkg/baselib/errors"
	"github.com/iGeeky/open-account/pkg/baselib/log"

	"golang.org/x/crypto/bcrypt"
)

// PwdEncode 使用bcrypt算法加密密码.
func PwdEncode(rawPassword string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
	return string(hashedBytes), err
}

// MustPwdEncode 使用bcrypt算法加密密码, 出错会抛出ERR_SERVER_ERROR异常.
func MustPwdEncode(rawPassword string) (encodePassword string) {
	var err error
	encodePassword, err = PwdEncode(rawPassword)
	if err != nil {
		log.Errorf("PwdEncode(%s) failed! err: %v", rawPassword, err)
		panic(errors.NewError(errors.ErrServerError, ""))
	}
	return
}

// PwdCompare 比较密码.
func PwdCompare(hashedPassword, rawPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(rawPassword))
}
