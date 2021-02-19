package net

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/iGeeky/open-account/pkg/baselib/errors"
	"github.com/iGeeky/open-account/pkg/baselib/log"
	"github.com/iGeeky/open-account/pkg/baselib/utils"
)

const (
	// ErrContentTypeInvalid 请求的Content-Type不合法.
	ErrContentTypeInvalid string = "ERR_CONTENT_TYPE_INVALID"
	// ErrOpenInputFile 打开文件出错
	ErrOpenInputFile string = "ERR_OPEN_INPUT_FILE"
)

// 需要签名的请求头前缀
var gCustomHeaderPrefix = "X-OA-"

// InitCustomHeaderPrefix 初始化
func InitCustomHeaderPrefix(customHeaderPrefix string) {
	if customHeaderPrefix != "" {
		gCustomHeaderPrefix = customHeaderPrefix
	}
}

// CustomHeaderName 获取自定义的请求头名.
func CustomHeaderName(headerName string) (customHeaderName string) {
	return gCustomHeaderPrefix + headerName
}

// [Content-Type] = {文件扩展名，路径前缀}
var gContentType map[string][]string
var gExtContentType map[string]string

func init() {
	gContentType = make(map[string][]string)
	// 图片
	gContentType["image/gif"] = []string{"gif", "img"}
	gContentType["image/jpeg"] = []string{"jpeg", "img"}
	gContentType["image/jpg"] = []string{"jpg", "img"}
	gContentType["image/png"] = []string{"png", "img"}
	gContentType["image/x-png"] = []string{"png", "img"}
	gContentType["image/x-png"] = []string{"png", "img"}
	gContentType["image/bmp"] = []string{"bmp", "img"}

	// 视频
	gContentType["video/mp4"] = []string{"mp4", "video"}
	gContentType["video/x-matroska"] = []string{"mkv", "video"}
	gContentType["video/x-msvideo"] = []string{"avi", "video"}
	gContentType["application/vnd.rn-realmedia-vbr"] = []string{"rmvb", "video"}
	gContentType["video/3gpp"] = []string{"3gp", "video"}
	gContentType["video/x-flv"] = []string{"flv", "video"}
	gContentType["video/mpeg"] = []string{"mpg", "video"}
	gContentType["video/quicktime"] = []string{"mov", "video"}
	gContentType["video/x-ms-wmv"] = []string{"wmv", "video"}
	gContentType["application/vnd.apple.mpegurl"] = []string{"m3u8", "video"}

	//音频
	gContentType["audio/wav"] = []string{"wav", "audio"}

	// 文本文件
	gContentType["text/plain"] = []string{"txt", "file"}
	gContentType["application/octet-stream"] = []string{"bin", "file"}
	gContentType["application/vnd.android.package-archive"] = []string{"apk", "file"}
	gContentType["application/iphone-package-archive"] = []string{"ipa", "file"}
	gContentType["text/xml"] = []string{"plist", "file"}

	gExtContentType = make(map[string]string)
	for contentType, arr := range gContentType {
		ext := arr[0]
		gExtContentType[ext] = contentType
	}
}

func getContentTypeByExt(ext string) string {
	return gExtContentType[ext]
}

func getContentType(filename string) string {
	ext := filepath.Ext(strings.ToLower(filename))
	if strings.HasPrefix(ext, ".") {
		ext = ext[1:]
	}
	contentType := getContentTypeByExt(ext)
	log.Infof("### filename: %s, ext: %s, contentType: %s", filename, ext, contentType)

	// if contentType == "" {
	// 	contentType = "application/octet-stream"
	// }

	return contentType
}

func CheckFileExist(host, filename, hash, appID, appKey, id string) *OkJson {
	uri := host + "/upload/check_exist"

	res := &OkJson{Ok: false, Reason: errors.ErrServerError}
	contentType := getContentType(filename)
	if contentType == "" {
		//fmt.Println("不支持的文件类型：", filename)
		log.Errorf("un support file type: %s", filename)
		res.ReqDebug = uri
		res.Reason = ErrContentTypeInvalid
		res.StatusCode = 400
		return res
	}

	headers := make(map[string]string, 10)
	headers["Content-Type"] = contentType
	headers[CustomHeaderName("Platform")] = "test"
	headers[CustomHeaderName("hash")] = hash
	headers[CustomHeaderName("AppID")] = appID
	if id != "" {
		headers[CustomHeaderName("ID")] = id
	}

	if appKey == "439081c403882a0c86fbed7ce2b4932cfcad47e1" {
		headers[CustomHeaderName("SIGN")] = appKey
		appKey = ""
	}

	res = HttpGetWithSign(uri, headers, 10, appKey)

	return res
}

func PostBody2UploadSvr(host, filename string, content []byte, appID, appKey, id, target, fileType, imageProcess string, timeout time.Duration, isTest bool) *OkJson {
	uri := host + "/upload/simple"
	res := &OkJson{Ok: false, Reason: errors.ErrServerError}

	contentType := getContentType(filename)
	if contentType == "" {
		log.Errorf("un support file type: %s", filename)
		res.Reason = ErrContentTypeInvalid
		res.StatusCode = 400
		return res
	}

	hash := utils.Sha1hex(content)
	headers := make(map[string]string, 10)
	headers["Content-Type"] = contentType
	headers[CustomHeaderName("Platform")] = "test"
	headers[CustomHeaderName("hash")] = hash
	headers[CustomHeaderName("AppID")] = appID
	if id != "" {
		headers[CustomHeaderName("ID")] = id
	}
	if fileType != "" {
		headers[CustomHeaderName("Type")] = fileType
	}
	if imageProcess != "" {
		headers[CustomHeaderName("ImageProcess")] = imageProcess
	}
	if target != "" {
		headers[CustomHeaderName("Target")] = target
	}
	if isTest {
		headers[CustomHeaderName("Test")] = "1"
	}

	if appKey == "439081c403882a0c86fbed7ce2b4932cfcad47e1" {
		headers[CustomHeaderName("SIGN")] = appKey
		appKey = ""
	}

	res = HttpPostWithSign(uri, content, headers, timeout, appKey)

	return res
}

func PostFile2UploadSvr(host, filename, appID, appKey, id, target, fileType, imageProcess string, timeout time.Duration, isTest bool) *OkJson {
	bodyfile, err := os.Open(filename)
	res := &OkJson{Ok: false, Reason: errors.ErrServerError}
	if err != nil {
		log.Errorf("Open file failed! err: %v", err)
		res.Reason = ErrOpenInputFile
		res.StatusCode = 400
		return res
	}
	defer bodyfile.Close()

	content, _ := ioutil.ReadFile(filename)

	res = PostBody2UploadSvr(host, filename, content, appID, appKey, id, target, fileType, imageProcess, timeout, isTest)
	return res
}

type UploadURLSimpleReq struct {
	URL             string `json:"url" validate:"required"`
	Referer         string `json:"referer"`
	UserAgent       string `json:"user_agent"`
	ContentType     string `json:"content_type"`
	ImageProcess    string `json:"imageProcess"`
	StackBlurRadius uint32 `json:"stack_blur_radius"`
}

func UploadURL(host, appID, appKey string, uploadURL *UploadURLSimpleReq, timeout time.Duration, isTest bool) *OkJson {
	uri := host + "/upload/url/simple"
	res := &OkJson{Ok: false, Reason: errors.ErrServerError}

	headers := make(map[string]string, 10)
	headers[CustomHeaderName("Platform")] = "test"
	headers[CustomHeaderName("AppID")] = appID
	if isTest {
		headers[CustomHeaderName("Test")] = "1"
	}

	if appKey == "439081c403882a0c86fbed7ce2b4932cfcad47e1" {
		headers[CustomHeaderName("SIGN")] = appKey
		appKey = ""
	}

	body, err := json.Marshal(uploadURL)
	if err != nil {
		res.Error = err
		return res
	}

	res = HttpPostWithSign(uri, body, headers, timeout, appKey)

	return res
}
