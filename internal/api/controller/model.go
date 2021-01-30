package controller

import "encoding/json"

// ManagerUserResetPwdReq 用户重置密码
type ManagerUserResetPwdReq struct {
	Tel      string `json:"tel" validate:"required,len=11"`
	Password string `json:"password" validate:"required"` // 密码md5后的值
	UserType int32  `json:"userType"`
}

// ManagerUserDeRegisterReq 删除用户.
type ManagerUserDeRegisterReq struct {
	ID int64 `json:"id"`
}

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

// UserSetInfoReq 设置用户基本信息.
type UserSetInfoReq struct {
	Avatar   string `json:"avatar"`   //头像URL
	Nickname string `json:"nickname"` //用户名
	Sex      int16  `json:"sex"`      //个人签名/简介
	Birthday string `json:"birthday"`
}

// UserSetNameReq 设置用户名(用于登录)
type UserSetNameReq struct {
	Username string `json:"username" validate:"required"` //用户名
}

// UserChangePasswordReq 用户修改密码
type UserChangePasswordReq struct {
	OldPassword string `json:"oldPassword" validate:"required,len=40"` // 旧的密码(hash后)
	Password    string `json:"password" validate:"required,len=40"`    // 密码sha1后的值
}

// UserResetPasswordReq 用户重置密码.
type UserResetPasswordReq struct {
	TelCodeInfo
	Password string `json:"password" validate:"required,len=40"` // 密码sha1后的值
	UserType int32  `json:"userType"`
}

// SetInviteCodeReq 设置注册邀请码
type SetInviteCodeReq struct {
	InviteCode string `json:"inviteCode" validate:"required"`
}
