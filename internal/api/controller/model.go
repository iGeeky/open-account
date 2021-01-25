package controller

import "encoding/json"

// UserSmsSendReq 用户发送短信请求结构
type UserSmsSendReq struct {
	BizType string `json:"bizType"`
	Tel     string `json:"tel" validate:"required,len=11"`
}

// TelCodeInfo 用户验证码基础结构
type TelCodeInfo struct {
	Tel  string `json:"tel" validate:"required,len=11"`
	Code string `json:"code" validate:"required"` // 验证码
}

// SmsLoginReq 短信验证登录/注册
type SmsLoginReq struct {
	TelCodeInfo
	UserType   int32  `json:"userType"`
	InviteCode string `json:"inviteCode"` // 邀请码.
}

// SMSCheckReq 校验验证码.
type SMSCheckReq struct {
	TelCodeInfo
	BizType string `json:"bizType"`
}

// UserRegisterReq 用户注册请求(手机号注册)
type UserRegisterReq struct {
	TelCodeInfo
	UserType   int32           `json:"userType"`
	Password   string          `json:"password"`   // 密码md5后的值
	Username   string          `json:"username"`   //用户名, 必须唯一
	Nickname   string          `json:"nickname"`   // 用户昵称, 可以重复
	InviteCode string          `json:"inviteCode"` // 邀请码.
	Profile    json.RawMessage `json:"profile"`
}

// UserLoginReq 用户密码登录请求.
type UserLoginReq struct {
	Tel      string `json:"tel" validate:"required,len=11"`
	Password string `json:"password" validate:"required,len=40"` // 密码md5后的值
	UserType int32  `json:"userType"`
}
