package ginplus

import (
	"bytes"
	"fmt"
	"net/url"
	"open-account/pkg/baselib/errors"
	"open-account/pkg/baselib/log"
	"open-account/pkg/baselib/utils"
	"regexp"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
)

// 需要签名的请求头
var defSignHeaders = []string{"host", "date"}
var signHeaderMap = map[string]bool{}

// 需要签名的请求头前缀
var gSignHeaderPrefix = "X-OA-"

type APISignConfig struct {
	Debug        bool
	CheckSign    bool
	DebugSignKey string
	AppKeys      map[string]string
	SignUrls     map[string]bool
}

// SignConfig global signature configuration
var SignConfig APISignConfig = APISignConfig{false, true, "", make(map[string]string), make(map[string]bool)}

func InitSign(signHeaders []string, signHeaderPrefix string) {
	if len(signHeaders) == 0 {
		signHeaders = defSignHeaders
	}
	for _, header := range signHeaders {
		signHeaderMap[strings.ToLower(header)] = true
	}
	if signHeaderPrefix != "" {
		gSignHeaderPrefix = signHeaderPrefix
	}
}

var pattern *regexp.Regexp = regexp.MustCompile("\\%[0-9A-Fa-f]{2}")

func IsEncoded(str string) bool {
	find := pattern.FindString(str)
	return find != ""
}

func uriEncodeInternal(arg string, encodeSlash bool) string {
	if arg == "" || IsEncoded(arg) {
		return arg
	}

	chars := bytes.NewBuffer([]byte{})

	barg := []byte(arg)
	for _, ch := range barg {
		if (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '_' || ch == '-' || ch == '~' || ch == '.' {
			chars.WriteByte(ch)
		} else if ch == '/' {
			if encodeSlash {
				chars.WriteString("%2F")
			} else {
				chars.WriteByte(ch)
			}
		} else {
			chars.WriteString(fmt.Sprintf("%%%02X", ch))
		}
	}

	return chars.String()
}

func URI_ENCODE(uri string) string {
	return uriEncodeInternal(uri, false)
}

func uri_encode(uri string) string {
	return uriEncodeInternal(uri, true)
}

func createCanonicalArgs(args map[string][]string) string {
	if args == nil {
		return ""
	}
	var keys []string

	for k, _ := range args {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	keyValues := []string{}

	for _, key := range keys {
		value := args[key]
		if len(value) == 1 {
			keyValues = append(keyValues, uri_encode(key)+"="+uri_encode(value[0]))
			// keyValues.WriteString()
		} else { // is array
			sort.Strings(value)
			for _, value_sub := range value {
				// keyValues.WriteString(uri_encode(key) + "=" + uri_encode(value_sub))
				keyValues = append(keyValues, uri_encode(key)+"="+uri_encode(value_sub))
			}
		}
	}

	return strings.Join(keyValues, "&")
}

// map[string][]string
func createCanonicalHeaders(headers map[string][]string) (string, string) {
	lowerHeaders := make(map[string]string)
	signedHeaders := []string{}
	ignoreSignHeader := gSignHeaderPrefix + "sign"
	for k, v := range headers {
		k = strings.ToLower(k)
		// TODO: value 是数组的情况。
		_, sign := signHeaderMap[k]
		if k != ignoreSignHeader && (sign || strings.HasPrefix(k, gSignHeaderPrefix)) {
			signedHeaders = append(signedHeaders, k)
			sort.Strings(v)
			lowerHeaders[k] = strings.Join(v, ",")
		}
	}

	headerValues := bytes.NewBuffer([]byte{})
	sort.Strings(signedHeaders)
	for i, k := range signedHeaders {
		if i != 0 {
			headerValues.WriteString("\n")
		}
		headerValues.WriteString(k + ":" + strings.TrimSpace(lowerHeaders[k]))
	}
	return headerValues.String(), strings.Join(signedHeaders, ";")
}

func createSignStr(uri string, args map[string][]string, headers map[string][]string, bodyHash, appKey string) string {

	CanonicalURI := URI_ENCODE(uri)
	CanonicalArgs := createCanonicalArgs(args)
	CanonicalHeaders, SignedHeaders := createCanonicalHeaders(headers)

	SignStr := CanonicalURI + "\n" +
		CanonicalArgs + "\n" +
		CanonicalHeaders + "\n" +
		SignedHeaders + "\n" +
		bodyHash + "\n" +
		appKey

	return SignStr
}

func GetUriPath(uri string) (string, error) {
	myurl, err := url.Parse(uri)
	if err != nil {
		return "", err
	}
	return myurl.Path, err
}

var EMPTY_BODY []byte = []byte("")

//哈希函数. 输入数据, 输出该数据的hex编码后的hash值.
type HexHashFunc func(data []byte) string

func SignEx(uri string, args map[string][]string, headers map[string][]string, body []byte, appKey string, hash HexHashFunc) (signature string, signStr string) {
	path, err := GetUriPath(uri)
	if err != nil {
		path = uri
	}

	if body == nil {
		body = EMPTY_BODY
	}
	bodyHash := hash(body)
	signStr = createSignStr(path, args, headers, bodyHash, appKey)
	btSignStr := []byte(signStr)
	signature = hash(btSignStr)

	return
}

func Sign(uri string, args map[string][]string, headers map[string][]string, body []byte, appKey string) (signature string, signStr string) {
	signature, signStr = SignEx(uri, args, headers, body, appKey, utils.Sha1hex)
	return
}

func SignSimple(uri string, args map[string]string, headers map[string]string, body []byte, appKey string) (signature string, signStr string) {
	myargs := make(map[string][]string)
	for k, v := range args {
		value := make([]string, 1)
		value[0] = v
		myargs[k] = value
	}

	myheaders := make(map[string][]string)
	for k, v := range headers {
		value := make([]string, 1)
		value[0] = v
		myheaders[k] = value
	}
	signature, signStr = Sign(uri, myargs, myheaders, body, appKey)
	return
}

// GetSignKey get signature key by appID
func (a *APISignConfig) GetSignKey(appID string) string {
	return a.AppKeys[appID]
}

// apiSignCheck check api signature
func apiSignCheck(ctx *ContextPlus, body []byte) bool {
	_ = ctx.Request.ParseForm()
	uri := ctx.Request.RequestURI
	args := ctx.Request.Form

	appID := ctx.GetCustomHeader("AppID")
	if appID == "" {
		log.Errorf("header '%s' missing.", ctx.CustomHeaderName("AppID"))
		ctx.JSON(401, gin.H{"ok": false, "reason": errors.ErrArgsInvalid})
		ctx.Abort()
		return false
	}
	ctx.Set("appID", appID)

	appKey := SignConfig.GetSignKey(appID)
	if appKey == "" {
		log.Errorf("Unknown appID [%s]", appID)
		ctx.JSON(401, gin.H{"ok": false, "reason": errors.ErrArgsInvalid})
		ctx.Abort()
		return false
	}

	reqSign := ctx.GetCustomHeader("Sign")
	if reqSign == "" {
		log.Errorf("header '%s' missing.", ctx.CustomHeaderName("Sign"))
		ctx.JSON(401, gin.H{"ok": false, "reason": errors.ErrSignError})
		ctx.Abort()
		return false
	}
	// log.Errorf("reqSign: %s", reqSign)
	// 测试工具使用。
	if SignConfig.DebugSignKey != "" && reqSign == SignConfig.DebugSignKey && SignConfig.Debug {
		return true
	}

	signature, SignStr := SignEx(uri, args, ctx.Request.Header, body, appKey, utils.Sha1hex)
	if signature != reqSign {
		log.Errorf("reqSign: [%s] != serverSign: [%s] \nSignStr [[%s]]", reqSign, signature, SignStr)
		log.Infof("req body len: %d", len(body))
		if SignConfig.Debug && len(body) < 100 {
			log.Infof("body: [[%v]]", string(body))
		}
		resp := gin.H{"ok": false, "reason": errors.ErrSignError}
		ctx.JSON(401, resp)
		ctx.Abort()
		return false
	}
	return true
}

// SignCheck check the signature
func SignCheck(c *gin.Context) {
	if c.Request.Method == "OPTIONS" {
		c.Next()
		return
	}
	ctx := NewContetPlus(c)
	uri := ctx.GetURI()

	if SignConfig.CheckSign && SignConfig.SignUrls[uri] {
		body, _ := ctx.GetBody()
		apiSignCheck(ctx, body)
	}

	c.Next()

}
