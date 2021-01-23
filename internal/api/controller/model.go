package controller

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
