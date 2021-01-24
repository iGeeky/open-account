package ginplus

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"open-account/pkg/baselib/errors"
	"open-account/pkg/baselib/log"
	"open-account/pkg/baselib/utils"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	valid "github.com/go-playground/validator/v10"
)

// ContextPlus 扩展的上下文.
type ContextPlus struct {
	*gin.Context
	body []byte
}

// NewContetPlus 创建一个扩展上下文.
func NewContetPlus(c *gin.Context) (context *ContextPlus) {
	context = &ContextPlus{
		Context: c,
	}
	return
}

// GetURI 获取URI, 不包含参数.
func (c *ContextPlus) GetURI() string {
	uri := c.Request.RequestURI

	pos := strings.Index(uri, "?")
	if pos >= 0 && pos < len(uri) {
		uri = uri[0:pos]
	}
	return uri
}

// GetBody 获取body内容, 多次调用时, 会使用缓存.
func (c *ContextPlus) GetBody() (body []byte, err error) {
	if c.body == nil {
		body = []byte("")
		defer utils.Elapsed("ReadBody")()
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "DELETE" {
			body, err = ioutil.ReadAll(c.Request.Body)
			if err != nil {
				return
			}
			c.Request.Body = ioutil.NopCloser(bytes.NewReader(body))
			c.body = body
		}
	} else {
		body = c.body
	}
	return
}

// ParseQueryJSONObject 解析请求体.
func (c *ContextPlus) ParseQueryJSONObject(query interface{}) {
	buf, err := c.GetBody()
	if err != nil {
		log.Errorf("Invalid body[%v] err: %v", string(buf), err)
		panic(errors.NewError(errors.ErrArgsInvalid, err.Error()))
	}

	log.Infof("uri:%v body:%v", c.GetURI(), string(buf))

	err = json.Unmarshal(buf, query)
	if err != nil {
		PostForm := c.Request.PostForm
		Form := c.Request.Form
		log.Errorf("Invalid body[%v] PostForm[%v] Form[%v], err: %v", string(buf), PostForm, Form, err)
		panic(errors.NewError(errors.ErrArgsInvalid, err.Error()))
	}

	validator := valid.New()
	if err = validator.Struct(query); err != nil {
		log.Errorf("Invalid Input Json [%s]! err: %s", string(buf), err)
		panic(errors.NewError(errors.ErrArgsInvalid, err.Error()))
	}
}

// Get 获取Get参数.
func (c *ContextPlus) Get(key string) (value string) {
	value = c.Query(key)
	return
}

// MustGet 获取Get参数, 如果为空,抛出ErrArgsMissing错误.
func (c *ContextPlus) MustGet(key string) (value string) {
	value = c.Get(key)
	if value == "" {
		log.Errorf("uri:%v missing args [%s]", c.GetURI(), key)
		panic(errors.NewError(errors.ErrArgsMissing, fmt.Sprintf("args [%s] is missing", key)))
	}
	return
}

// GetBool 获取bool值, 其中"",true, yes, on, y, t 都表示真, 其它值都表示假.
func (c *ContextPlus) GetBool(key string) (b bool) {
	value, exist := c.GetQuery(key)
	if !exist {
		return
	}

	b = value == "" || value == "true" || value == "yes" || value == "on" || value == "y" || value == "t"

	return
}

// GetBoolPtr 获取bool值, 其中"",true, yes, on, y, t 都表示真, 其它值都表示假. 与GetBool不同的是, 当参数不存在时, 返回nil.
func (c *ContextPlus) GetBoolPtr(key string) (b *bool) {
	value, exist := c.GetQuery(key)
	if !exist {
		return
	}

	bvalue := value == "" || value == "true" || value == "yes" || value == "on" || value == "y" || value == "t"
	b = &bvalue

	return
}

// GetInt32 获取int32值. 如果为空或解析失败, 返回默认值
func (c *ContextPlus) GetInt32(key string, def int32) (value int32) {
	strvalue := c.Query(key)
	if strvalue == "" {
		value = def
		return
	}

	intValue, err := strconv.ParseInt(strvalue, 10, 32)
	if err != nil {
		log.Errorf("Invalid int32 value(%s) err: %v", strvalue, err)
		value = def
	} else {
		value = int32(intValue)
	}

	return
}

// GetInt64 获取int64值. 如果为空或解析失败, 返回默认值
func (c *ContextPlus) GetInt64(key string, def int64) (value int64) {
	strvalue := c.Query(key)
	if strvalue == "" {
		value = def
		return
	}
	intValue, err := strconv.ParseInt(strvalue, 10, 64)
	if err != nil {
		log.Errorf("Invalid int64 value(%s) err: %v", strvalue, err)
		value = def
	} else {
		value = intValue
	}

	return
}

// MustGetInt32 获取int32值. 如果出错,抛出ErrArgsInvalid错误.
func (c *ContextPlus) MustGetInt32(key string) (value int32) {
	strvalue := c.MustGet(key)
	intValue, err := strconv.ParseInt(strvalue, 10, 32)
	if err != nil {
		log.Errorf("uri:%v args [%s] value(%s) is invalid, need int32", c.GetURI(), key, strvalue)
		panic(errors.NewError(errors.ErrArgsInvalid, fmt.Sprintf("args [%s] value(%s) is invalid, need int32", key, strvalue)))
	}
	value = int32(intValue)
	return
}

// MustGetInt64 获取int64值. 如果出错,抛出ErrArgsInvalid错误.
func (c *ContextPlus) MustGetInt64(key string) (value int64) {
	strvalue := c.MustGet(key)
	intValue, err := strconv.ParseInt(strvalue, 10, 64)
	if err != nil {
		log.Errorf("uri:%v args [%s] value(%s) is invalid, need int64", c.GetURI(), key, strvalue)
		panic(errors.NewError(errors.ErrArgsInvalid, fmt.Sprintf("args [%s] value(%s) is invalid, need int64", key, strvalue)))
	}
	value = intValue
	return
}

// GetClientMetaInfo 获取请求头中的Platform,Version, Channel, deviceID 相关信息.
func (c *ContextPlus) GetClientMetaInfo() (platform, version, channel, deviceID string) {
	platform = c.GetHeader("X-OA-Platform")
	version = c.GetHeader("X-OA-Version")
	channel = c.GetHeader("X-OA-Channel")
	deviceID = c.GetHeader("X-OA-DeviceID")

	return
}

// JsonOk 返回成功及data节点
func (c *ContextPlus) JsonOk(data interface{}) gin.H {
	response := gin.H{"ok": true, "reason": "", "data": data}
	c.JSON(200, response)
	return response
}

// JsonOk 返回成功及reason, data节点
func (c *ContextPlus) JsonOk2(reason string, data interface{}) gin.H {
	response := gin.H{"ok": true, "reason": reason, "data": data}
	c.JSON(200, response)
	return response
}

// JsonFail 返回失败及reason节点
func (c *ContextPlus) JsonFail(reason string) gin.H {
	response := gin.H{"ok": false, "reason": reason}
	if reason == errors.ErrServerError {
		c.JSON(500, response)
	} else {
		c.JSON(200, response)
	}
	return response
}

// JsonFailWithMsg 返回失败及reason,errmsg节点
func (c *ContextPlus) JsonFailWithMsg(reason string, errmsg interface{}) gin.H {
	response := gin.H{"ok": false, "reason": reason, "errmsg": errmsg}
	if reason == errors.ErrServerError {
		c.JSON(500, response)
	} else {
		c.JSON(200, response)
	}
	return response
}
