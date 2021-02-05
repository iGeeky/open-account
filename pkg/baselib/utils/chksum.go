package utils

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"github.com/iGeeky/open-account/pkg/baselib/log"
)

var ErrDataInvalid error

func init() {
	ErrDataInvalid = fmt.Errorf("ERR_DATA_INVALID")
}

type CheckSum struct {
	magic  string //计算时添加的magic
	length int    // 校验码长度.
	digest string //编码方法: base64, hex
}

func NewCheckSum(magic string, length int, digest string) (checkSum *CheckSum) {
	if length < 1 {
		length = 3
	}
	checkSum = &CheckSum{magic: magic, length: length, digest: digest}
	return
}

func (c *CheckSum) calcChkSum(data string) (cksum string) {
	sum := md5.Sum([]byte(data + c.magic))
	var digestedSum string
	if c.digest == "base64" {
		digestedSum = base64.RawURLEncoding.EncodeToString(sum[:])
	} else {
		digestedSum = hex.EncodeToString(sum[:])
	}
	cksum = string(digestedSum[0:c.length])
	// fmt.Printf("data: %s, cksum [%s]\n", data, cksum)
	return
}

func (c *CheckSum) Add(data string) (newData string) {
	newData = data + c.calcChkSum(data)
	return
}

func (c *CheckSum) Check(data string) (rawData string, err error) {
	length := len(data)
	if length <= c.length+1 {
		err = ErrDataInvalid
		return
	}

	rawData = data[0 : length-c.length]
	cksum := data[length-c.length:]
	cksumCalc := c.calcChkSum(rawData)
	if cksum != cksumCalc {
		log.Errorf("invalid data [%s], the ok cksum is :%s", data, cksumCalc)
		err = ErrDataInvalid
		rawData = ""
		return
	}

	return
}
