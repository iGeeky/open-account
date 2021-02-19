package net

import (
	"fmt"
	"net/url"
	"time"

	"github.com/iGeeky/open-account/pkg/baselib/ginplus"
)

func GetContentType(filename string) string {
	return getContentType(filename)
}

func ParseArgs(uri string) (args map[string]string) {
	u, err := url.Parse(uri)
	if err != nil {
		return
	}
	values, err := url.ParseQuery(u.RawQuery)
	if err == nil {
		args = make(map[string]string)
		for k, arrValue := range values {
			args[k] = arrValue[0] // 有多个参数的, 只取第一个.所以请不要传入多个相同的参数, 会导致签名错误.
		}
		return
	}
	return
}

func HttpPostWithSign(uri string, body []byte, headers map[string]string, timeout time.Duration, appKey string) *OkJson {
	signature, SignStr := "", ""
	if appKey != "" {
		args := ParseArgs(uri)
		signature, SignStr = ginplus.SignSimple(uri, args, headers, body, appKey)
		headers[CustomHeaderName("SIGN")] = signature
	}

	res := HttpPostJson(uri, body, headers, timeout)

	if res.StatusCode == 401 {
		fmt.Printf("signature [%s] SigStr [[\n%s\n]]", signature, SignStr)
	}
	return res
}

func HttpGetWithSign(uri string, headers map[string]string, timeout time.Duration, appKey string) *OkJson {
	signature, SignStr := "", ""
	if appKey != "" {
		args := ParseArgs(uri)
		signature, SignStr = ginplus.SignSimple(uri, args, headers, nil, appKey)
		headers[CustomHeaderName("SIGN")] = signature
	}

	res := HttpGetJson(uri, headers, timeout)

	if res.StatusCode == 401 {
		fmt.Printf("signature [%s] SigStr [[\n%s\n]]", signature, SignStr)
	}
	return res
}
