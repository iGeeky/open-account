package utils

import (
	"encoding/base64"
	"encoding/binary"
	"math/rand"
	"open-account/pkg/baselib/log"
	"strconv"
	"time"
)

// Elapsed 计算函数执行时间.
func Elapsed(what string) func() {
	start := time.Now()
	return func() {
		log.Infof(">>>> %s took %v", what, time.Since(start))
	}
}

const float64EqualityThreshold = 0.000001

// Uint32ToBase64 uint32转成base64格式.
func Uint32ToBase64(value uint32) (base64val string) {
	src := make([]byte, 4)
	binary.LittleEndian.PutUint32(src, value)
	base64val = base64.RawURLEncoding.EncodeToString(src)
	return
}

// IsInteger 是否是integer数据
func IsInteger(str string) (ok bool) {
	_, err := strconv.ParseInt(str, 10, 64)
	ok = err == nil
	return
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// RandomString 随机指定长度的字符串
func RandomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

const codeLetters = "0123456789"

// RandomCode 随机指定长度的数字字符串
func RandomCode(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = codeLetters[rand.Intn(len(codeLetters))]
	}
	return string(b)
}

// RandomInviteCode 生成邀请码.
func RandomInviteCode(length int) string {
	return RandomCode(length)
}

// Now 当前时间,unix时间戳
func Now() (now int64) {
	now = time.Now().Unix()
	return
}

// CurrentDate 当前日期
func CurrentDate() (date string) {
	// format: 2006-01-02
	date = time.Now().Format("2006-01-02")
	return
}

// Datetime 当前时间
func Datetime() (datetime string) {
	// format: 2006-01-02 15:04:05
	datetime = time.Now().Format("2006-01-02 15:04:05")
	return
}

// UnixTimestamp 时间转时间戳.
func UnixTimestamp(datetime string) (unixtime int64) {
	t, err := time.Parse("2006-01-02 15:04:05", datetime)
	if err == nil {
		unixtime = t.Unix()
	}
	return
}
