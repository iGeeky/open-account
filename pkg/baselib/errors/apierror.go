package errors

// Assert 如果err不为空, 抛出ERR_SERVER_ERROR错误.
func Assert(err error) {
	if err != nil {
		panic(NewError(ErrServerError, err.Error()))
	}
}

const (
	// ErrArgsInvalid 参数不合法
	ErrArgsInvalid string = "ERR_ARGS_INVALID"
	// ErrArgsMissing 参数缺失
	ErrArgsMissing string = "ERR_ARGS_MISSING"
	// ErrServerError 服务器错误
	ErrServerError string = "ERR_SERVER_ERROR"
	// ErrSignError 签名错误
	ErrSignError string = "ERR_SIGN_ERROR"
	// ErrTokenInvalid Token不合法
	ErrTokenInvalid string = "ERR_TOKEN_INVALID"
	// ErrTokenExpired Token已经过期
	ErrTokenExpired string = "ERR_TOKEN_EXPIRED"
	// ErrTokenOfflined Token已经掉线(被踢掉)
	ErrTokenOfflined string = "ERR_TOKEN_OFFLINED"
	// ErrObjectNotFound 查询的对象不存在
	ErrObjectNotFound string = "ERR_OBJECT_NOT_FOUND"
	// ErrObjectExist 要添加的对象已经存在
	ErrObjectExist string = "ERR_OBJECT_EXIST"
	// ErrCodeInvalid 验证码错误
	ErrCodeInvalid string = "ERR_CODE_INVALID"
	// ErrUserIsLocked 用户被锁定, 不能登录.
	ErrUserIsLocked string = "ERR_USER_IS_LOCKED"
	// ErrTelRegisted 手机号已经注册过了.
	ErrTelRegisted string = "ERR_TEL_REGISTED"
	// ErrUsernameRegisted 用户名已经注册过了
	ErrUsernameRegisted string = "ERR_USERNAME_REGISTED"
	// ErrTelNotExist 手机不存在.
	ErrTelNotExist string = "ERR_TEL_NOT_EXIST"
	// ErrUsernameNotExist 用户名不存在
	ErrUsernameNotExist string = "ERR_USERNAME_NOT_EXIST"
	// ErrPasswordErr 密码错误
	ErrPasswordErr string = "ERR_PASSWORD_ERR"
	// ErrInviteCodeInvalid 邀请码不合法
	ErrInviteCodeInvalid string = "ERR_INVITE_CODE_INVALID"
	// ErrUserHaveUsername 用户已经有用户名了, 不能再设置.
	ErrUserHaveUsername string = "ERR_USER_HAVE_USERNAME"
	// ErrCaptchaError 人机验证码错误
	ErrCaptchaError string = "ERR_CAPTCHA_ERROR"
)

// GetStatusCode 获取错误对应的响应码
func GetStatusCode(reason string) int {
	switch reason {
	case ErrArgsInvalid, ErrArgsMissing:
		return 400
	case ErrSignError, ErrTokenInvalid, ErrTokenExpired, ErrTokenOfflined:
		return 401
	case ErrServerError:
		return 500
	default:
		// log.Errorf("Unknown Reason: %s", reason)
		return 200
	}
}

type ApiError struct {
	Reason string `json:"reason"`
	Errmsg string `json:"errmsg,omitempty"`
}

func NewError(reason string, errmsg string) (err *ApiError) {
	err = &ApiError{Reason: reason, Errmsg: errmsg}
	return
}
