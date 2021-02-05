package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"github.com/iGeeky/open-account/pkg/baselib/log"
	"github.com/iGeeky/open-account/pkg/baselib/utils"
)

// type TokenInfo struct {
// 	Token      string
// 	UserType   int16
// 	UserID     int64
// 	ExpireTime int64
// }

var tokenBlock cipher.Block

var tokenKey []byte
var tokenIV []byte
var tokenMagic string
var asciiTable string

var tokenV1Prefix string          //注册用户token
var anonymousTokenV1Prefix string //匿名用户

var errTokenInvalid error
var tokenChksum *utils.CheckSum

func init() {
	var err error
	tokenKey = []byte("cde6ca958b95b08db5f53c8f583dcb62")
	tokenIV = []byte("6760ca184fcee8c3194bf377d213d900")
	tokenMagic = "1654fb0cf72bdcaf"[:16]
	asciiTable = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	tokenV1Prefix = "T1"
	anonymousTokenV1Prefix = "A1"
	tokenChksum = utils.NewCheckSum(tokenMagic, 4, "base64")

	tokenBlock, err = aes.NewCipher(tokenKey)
	if err != nil {
		log.Errorf("NewCipher(%s) failed! err:%v", tokenKey, err)
		panic(err.Error())
	}
	tokenIV = tokenIV[:tokenBlock.BlockSize()]

	errTokenInvalid = fmt.Errorf("ERR_TOKEN_INVALID")
}

func randomSID() string {
	return utils.RandomString(8)
}

func tokenEncryptInternal(prefix string, userType int32, userID int64, expireTime int64) string {
	tokenTmp := fmt.Sprintf("%x|%x|%x|%s", userType, userID, expireTime, randomSID())
	src := []byte(tokenTmp)
	dst := make([]byte, len(src))
	encrypter := cipher.NewCFBEncrypter(tokenBlock, tokenIV)
	encrypter.XORKeyStream(dst, src)
	tokenTmp = base64.RawStdEncoding.EncodeToString(dst)
	return tokenChksum.Add(prefix + tokenTmp)
}

// TokenEncrypt token加密
func TokenEncrypt(userType int32, userID int64, expireTime int64) string {
	return tokenEncryptInternal(tokenV1Prefix, userType, userID, expireTime)
}

// AnonymousTokenEncrypt 匿名用户token加密.
func AnonymousTokenEncrypt(userType int32, userID int64, expireTime int64) string {
	return tokenEncryptInternal(anonymousTokenV1Prefix, userType, userID, expireTime)
}

// TokenDecrypt token解密
func TokenDecrypt(token string) (userType int32, userID int64, expireTime int64, err error) {
	ver := token[0:2]
	if ver == tokenV1Prefix || ver == anonymousTokenV1Prefix {
		var rawToken string
		rawToken, err = tokenChksum.Check(token)
		if err != nil {
			err = errTokenInvalid
			return
		}
		rawToken = rawToken[2:]
		var rawTokenBts []byte
		rawTokenBts, err = base64.RawStdEncoding.DecodeString(rawToken)
		if err != nil {
			log.Errorf("base64.DecodeString(%s) failed! err:%v", rawToken, err)
			err = errTokenInvalid
			return
		}
		dst := make([]byte, len(rawTokenBts))
		decrypter := cipher.NewCFBDecrypter(tokenBlock, tokenIV)
		decrypter.XORKeyStream(dst, rawTokenBts)
		rawToken = string(dst)
		arr := strings.SplitN(rawToken, "|", 4)
		if len(arr) != 4 {
			log.Errorf("invalid token: %s", rawToken)
			err = errTokenInvalid
			return
		}
		var i64UserType int64
		i64UserType, err = strconv.ParseInt(arr[0], 16, 16)
		if err != nil {
			return
		}
		userType = int32(i64UserType)
		userID, err = strconv.ParseInt(arr[1], 16, 64)
		if err != nil {
			return
		}
		expireTime, err = strconv.ParseInt(arr[2], 16, 64)
		if err != nil {
			return
		}
	} else {
		err = errTokenInvalid
	}

	return
}
