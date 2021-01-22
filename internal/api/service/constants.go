package service

const (
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
