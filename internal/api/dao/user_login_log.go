package dao

// UserLoginLog 用户登录日志
type UserLoginLog struct {
	ID          int64  `json:"id" gorm:"primary_key;AUTO_INCREMENT"`
	UserID      int64  `json:"userID"`
	DeviceID    string `json:"deviceID"`
	LoginIP     string `json:"loginIP"`
	CountryCode string `json:"countryCode"`
	CityName    string `json:"cityName"`
	Channel     string `json:"channel"`
	Platform    string `json:"platform"`
	Version     string `json:"version"`
	CreateTime  int64  `json:"createTime"`
}

// TableName 表名
func (u *UserLoginLog) TableName() string {
	return "user_login_log"
}

// UserLoginLogDao 用户登录日志Dao
type UserLoginLogDao struct {
	*AccountBaseDao
}

// NewUserLoginLogDao 创建新的UserLoginLogDao
func NewUserLoginLogDao() (dao *UserLoginLogDao) {
	dao = &UserLoginLogDao{AccountBaseDao: NewAccountBaseDao(&UserInfo{})}
	return
}
