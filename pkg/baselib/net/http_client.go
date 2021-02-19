package net

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/iGeeky/open-account/pkg/baselib/log"
)

func init() {
}

const (
	ErrArgsInvalid string = "ERR_ARGS_INVALID"
	ErrServerError string = "ERR_SERVER_ERROR"
)

type ReqStats struct {
	All time.Duration //请求总时间。
}

type Resp struct {
	Error      error       // 出错信息。
	RawBody    []byte      // Http返回的原始内容。
	StatusCode int         // Http响应吗
	Headers    http.Header // HTTP响应头
	ReqDebug   string      // 请求的DEBUG串(curl格式)
	Stats      ReqStats    //请求时间统计。
}

type OkJson struct {
	*Resp
	Ok     bool                   `json:"ok"`
	Reason string                 `json:"reason"`
	Data   map[string]interface{} `json:"data"`
	Cached string                 //数据是否是缓存中获取 hit, miss
}

func (o *OkJson) Body() string {
	return string(o.RawBody)
}

func headerstr(headers map[string]string) string {
	if headers == nil {
		return ""
	}

	lines := make([]string, 4)
	for k, v := range headers {
		// if k != "User-Agent" {
		if strings.TrimSpace(v) == "" {
			lines = append(lines, fmt.Sprintf("-H'%s;'", k))
		} else {
			lines = append(lines, fmt.Sprintf("-H'%s: %s'", k, v))
		}
		// }
	}

	return strings.Join(lines, " ")
}

var gProxyURL string

func HttpHeaderConvert(srcHeaders http.Header) (headers map[string]string) {
	headers = make(map[string]string)
	if srcHeaders != nil {
		for k, arrValue := range srcHeaders {
			headers[k] = arrValue[0]
		}
	}
	return
}

// proxyURL : "http://" + p.AppID + ":" + p.AppSecret + "@" + ProxyServer
func SetProxyURL(proxyURL string) {
	gProxyURL = proxyURL
}

func isTextContent(contentType string, headers map[string]string) bool {
	ok := contentType == "" || strings.HasPrefix(contentType, "text") ||
		strings.HasPrefix(contentType, "application/json") ||
		strings.HasPrefix(contentType, "application/x-www-form-urlencoded") ||
		headers["X-Body-Is-Text"] == "1"
	return ok
}

func HttpReqDebug(method, uri string, body []byte, headers map[string]string, maxBodyLen int) string {
	var reqDebug string
	if method == "PUT" || method == "POST" {
		var debugBody string
		contentType := headers["Content-Type"]
		if isTextContent(contentType, headers) {
			if maxBodyLen == 0 || len(body) < maxBodyLen {
				debugBody = string(body)
			} else {
				debugBody = string(body[0:maxBodyLen])
			}
		} else {
			debugBody = "[[not text body: " + contentType + "]]"
		}
		reqDebug = "curl -v -X " + method + " " + headerstr(headers) + " '" + uri + "' -d '" + debugBody + "' "
	} else {
		reqDebug = "curl -v -X " + method + " " + headerstr(headers) + " '" + uri + "' "
	}
	return reqDebug
}

//支持原生调用，wangyanglong@nicefilm.com
//外部需要自己回收资源 defer resp.body.Close()
//https://golang.org/src/net/http/client.go
// The Client's Transport typically has internal state (cached TCP
// connections), so Clients should be reused instead of created as
// needed. Clients are safe for concurrent use by multiple goroutines.
var gDefaultClient *http.Client
var onceDefaultClient sync.Once
var onceProxyClient sync.Once

type proxyClientSet struct {
	sync.RWMutex
	clientMap map[string]*http.Client
}

var gProxyClient *proxyClientSet

func newHttpClient(transport *http.Transport, timeout time.Duration) *http.Client {
	client := &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
	return client
}

const (
	MaxIdleConnsPerHost = 15
	MaxIdleConns        = 500
)

func defaultHttpTransport() *http.Transport {
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          MaxIdleConns,
		MaxIdleConnsPerHost:   MaxIdleConnsPerHost,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 5 * time.Second,
		ResponseHeaderTimeout: 90 * time.Second,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
	}
	return transport
}

func NewHttpTransport(timeout time.Duration) *http.Transport {
	if timeout == 0 {
		return defaultHttpTransport()
	}

	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   timeout,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          MaxIdleConns,
		MaxIdleConnsPerHost:   MaxIdleConnsPerHost,
		IdleConnTimeout:       timeout,
		TLSHandshakeTimeout:   timeout,
		ExpectContinueTimeout: timeout,
		ResponseHeaderTimeout: timeout,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
	}
	return transport
}

func NewHttpClient(timeout time.Duration) (client *http.Client, err error) {
	onceDefaultClient.Do(func() {
		gDefaultClient = newHttpClient(NewHttpTransport(timeout), timeout)
	})
	if gDefaultClient != nil {
		client = gDefaultClient
	} else {
		err = errors.New("get default http client error,init failed")
	}
	return
}

func httpReq(method, uri string, body []byte, args map[string]string, headers map[string]string, timeout time.Duration) (*http.Response, error, string) {
	client, err := NewHttpClient(timeout)
	if err != nil {
		return nil, err, ""
	}

	req, err, _ := FormatHttpRequest(method, uri, args, headers, body)
	if err != nil {
		return nil, err, ""
	}
	urlWithArgs := req.URL.String()
	reqDebug := HttpReqDebug(method, urlWithArgs, body, headers, 1024)

	log.Infof("REQUEST [ %s ] timeout: %v", reqDebug, timeout)
	begin := time.Now()
	resp, err := client.Do(req) //发送
	cost := time.Now().Sub(begin)
	code := -1
	if resp != nil {
		code = resp.StatusCode
	}
	log.Infof("REQUEST [ %s ] status: %d, recv header cost: %v", reqDebug, code, cost)

	return resp, err, reqDebug
}

func FormatHttpRequest(method, uri string, args, headers map[string]string, body []byte) (req *http.Request, err error, reason string) {
	req, err = http.NewRequest(method, uri, bytes.NewReader(body))
	if err != nil {
		log.Errorf("NewRequest(method:%s, uri: '%s') failed! err: %v", method, uri, err)
		return nil, err, ""
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	q := req.URL.Query()
	for key, value := range args {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()
	req.Host = headers["Host"]
	return req, nil, ""
}

func writeDebug(begin time.Time, res *Resp, bodyLen *int) {
	cost := time.Now().Sub(begin)
	res.Stats.All = cost
	seconds := cost.Seconds()
	kbps := float64(0)
	if seconds > 0 && *bodyLen > 0 {
		kbps = float64(*bodyLen) / float64(1024) / seconds
	}
	log.Infof("REQUEST [ %s ] status: %d, bodyLen: %d, cost: %v, speed: %.3f kb/s", res.ReqDebug, res.StatusCode, *bodyLen, cost, kbps)
}

func writeDebugOKJson(begin time.Time, res *OkJson, bodyLen *int) {
	cost := time.Now().Sub(begin)
	res.Stats.All = cost
	seconds := cost.Seconds()
	kbps := float64(0)
	if seconds > 0 && *bodyLen > 0 {
		kbps = float64(*bodyLen) / float64(1024) / seconds
	}
	log.Infof("REQUEST [ %s ] status: %d, bodyLen: %d, cost: %v, speed: %.3f kb/s", res.ReqDebug, res.StatusCode, *bodyLen, cost, kbps)
}

func HttpReqInternal(method, uri string, body []byte, args map[string]string, headers map[string]string, timeout time.Duration) *Resp {
	res := &Resp{}
	begin := time.Now()
	bodyLen := 0

	resp, err, reqDebug := httpReq(method, uri, body, args, headers, timeout)
	res.ReqDebug = reqDebug
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close() //一定要关闭resp.Body
	}
	defer writeDebug(begin, res, &bodyLen)
	if err != nil {
		log.Errorf("###### err: %v", err)
		res.Error = err
		res.StatusCode = 500
		return res
	}

	res.Headers = resp.Header
	// defer resp.Body.Close() //一定要关闭resp.Body
	data, err := ioutil.ReadAll(resp.Body)
	if data != nil {
		bodyLen = len(data)
	}
	if err != nil {
		log.Errorf("REQUEST [ %s ] Read Body Failed! err: %v", reqDebug, err)
		res.StatusCode = 500
		res.Error = err
		return res
	}
	res.RawBody = data
	res.StatusCode = resp.StatusCode
	contentLength := 0
	strContentLengths := res.Headers["Content-Length"]
	if len(strContentLengths) > 0 {
		contentLength, _ = strconv.Atoi(strContentLengths[0])
		if contentLength > 0 && len(data) != contentLength {
			res.StatusCode = 500
			log.Errorf("REQUEST [ %s ] Content-Length: %d, len(body): %d", res.ReqDebug, contentLength, len(data))
			res.Error = errors.New("Length of Body is Invalid")
			return res
		}
	}
	if err != nil {
		log.Errorf("REQUEST [ %s ] Read Body Failed! body-len: %d err: %v", reqDebug, contentLength, err)
		res.StatusCode = 500
		res.Error = err
		return res
	}

	return res
}

func HttpGet(uri string, headers map[string]string, timeout time.Duration) *Resp {
	return HttpReqInternal("GET", uri, nil, nil, headers, timeout)
}

func HttpPost(uri string, body []byte, headers map[string]string, timeout time.Duration) *Resp {
	return HttpReqInternal("POST", uri, body, nil, headers, timeout)
}

func OkJsonParse(res *OkJson) *OkJson {
	decoder := json.NewDecoder(bytes.NewBuffer(res.RawBody))
	decoder.UseNumber()
	err := decoder.Decode(&res)
	if err != nil {
		log.Errorf("Invalid json [%s] err: %v", string(res.RawBody), err)
		res.Error = err
		res.Reason = ErrServerError
		res.StatusCode = 500
		return res
	}
	if !res.Ok && res.Reason != "" && res.Error == nil {
		res.Error = fmt.Errorf(res.Reason)
	}

	return res
}

func HttpReqJson(method, uri string, body []byte, args map[string]string, headers map[string]string, timeout time.Duration) *OkJson {

	resHttp := HttpReqInternal(method, uri, body, args, headers, timeout)
	res := &OkJson{Ok: false, Reason: ErrServerError, Resp: resHttp}
	// res.Error = resHttp.Error
	// res.RawBody = resHttp.RawBody
	// res.StatusCode = resHttp.StatusCode
	// res.Headers = resHttp.Headers
	// res.ReqDebug = resHttp.ReqDebug
	// res.Stats = resHttp.Stats

	if resHttp.StatusCode >= 500 {
		res.Reason = ErrServerError
		return res
	}

	OkJsonParse(res)

	return res
}

func HttpGetJson(uri string, headers map[string]string, timeout time.Duration) *OkJson {
	return HttpReqJson("GET", uri, nil, nil, headers, timeout)
}

func HttpPostJson(uri string, body []byte, headers map[string]string, timeout time.Duration) *OkJson {
	return HttpReqJson("POST", uri, body, nil, headers, timeout)
}
