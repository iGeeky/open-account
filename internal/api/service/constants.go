package service

const (
	TokenNone  = 0 //不验证Token
	TokenUser  = 1 //验证客户端登录Token
	TokenAdmin = 2 //管理后台使用配置的超级Token.

	UserSexUnknown   = int16(0)
	UserSexMale      = int16(1)
	UserSexFemale    = int16(2)
	UserSexOtherwise = int16(3)

	// normal user
	UserTypeNormal = int32(1)
	// 账号正常.
	UserStatusOK = int16(1)
	// 账号被禁用/锁定
	UserStatusDisabled = int16(-1)
)
