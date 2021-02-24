package net

import (
	"bytes"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/iGeeky/open-account/pkg/baselib/cache"
	"github.com/iGeeky/open-account/pkg/baselib/ginplus"
	"github.com/iGeeky/open-account/pkg/baselib/log"
)

var ReqDebugHost string
var ReqDebugDir string
var Debug bool

type reqDebugWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w reqDebugWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func init() {
	ReqDebugInit("http://127.0.0.1", "log", time.Minute*10, false)
}

func ReqDebugInit(reqDebugHost string, reqDebugDir string, fdExpireTime time.Duration, debug bool) {
	Debug = debug
	if Debug {
		ReqDebugHost = reqDebugHost
		ReqDebugDir = reqDebugDir
		cache.StartFileCleanupTimer(fdExpireTime)
	}
}

func isTextContent2(c *gin.Context) bool {
	contentType := c.GetHeader("Content-Type")
	ok := contentType == "" || strings.HasPrefix(contentType, "text") ||
		strings.HasPrefix(contentType, "application/json")
	return ok
}

func ReqDebug(c *gin.Context) {
	//上传的请求, 直接忽略.
	if !Debug || !isTextContent2(c) {
		c.Next()
		return
	}
	ctx := ginplus.NewContetPlus(c)

	rdw := &reqDebugWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
	c.Writer = rdw

	fullURI := ReqDebugHost + c.Request.RequestURI
	if !strings.HasPrefix(fullURI, "http://") {
		fullURI = "http://" + fullURI
	}
	headers := HttpHeaderConvert(c.Request.Header)
	delete(headers, "Content-Length")
	var body []byte
	var err error
	if c.Request.Method == "POST" || c.Request.Method == "PUT" {
		body, err = ctx.GetBody()
		if err != nil {
			log.Errorf("ctx.GetBody() failed! err: %+v", err)
		}
	}
	now := time.Now()
	filename := fmt.Sprintf("%02d%02d_%02d_req_debug.log", now.Month(), now.Day(), now.Hour())
	prefix := fmt.Sprintf("%02d:%02d:%02d: ", now.Hour(), now.Minute(), now.Second())

	logfilename := path.Join(ReqDebugDir, filename)

	reqDebug := HttpReqDebug(c.Request.Method, fullURI, body, headers, 0)
	if ReqDebugDir == "log" {
		log.Infof("Begin Request [%s] ...", reqDebug)
	} else {
		cache.BizLogWrite(logfilename, "REQ: "+prefix+reqDebug)
	}
	c.Next()

	rspBody := rdw.body.String()
	if ReqDebugDir == "log" {
		log.Infof("Response [%s]", rspBody)
	} else {
		cache.BizLogWrite(logfilename, "RSP: "+prefix+rspBody)
	}

	//context.Clear(c.Request)
}
