package dao

import (
	// "github.com/lib/pq"
	"database/sql/driver"
	"encoding/json"
	"open-account/pkg/baselib/db"
	"open-account/pkg/baselib/log"
	"open-account/pkg/baselib/utils"
)

// Profile 个人信息, 需要先在Profile类中定义相应的字段.
type Profile struct {
	ID   string `json:"id"`
	City string `json:"city"`
}

// Value 实现方法
func (p Profile) Value() (driver.Value, error) {
	return json.Marshal(p)
}

// Scan 实现方法
func (p *Profile) Scan(input interface{}) (err error) {
	buf := input.([]byte)
	err = json.Unmarshal(buf, p)
	return
}

// UserInfo 用户基本信息表
type UserInfo struct {
	ID            int64   `view:"detail,man" json:"id" gorm:"primary_key;AUTO_INCREMENT"`
	UID           string  `view:"*" json:"uid"`
	Tel           string  `view:"detail,man" json:"tel"`
	Password      string  `view:"-" json:"password"`
	Username      string  `view:"-" json:"username"`
	Nickname      string  `view:"*" json:"nickname"`
	Avatar        string  `view:"*" json:"avatar"`
	Sex           int16   `view:"*" json:"sex"`
	Birthday      string  `view:"*" json:"birthday"`
	UserType      int32   `view:"*" json:"userType"`
	RegInviteCode string  `view:"detail,man" json:"regInviteCode"`
	InviteCode    string  `view:"detail,man" json:"inviteCode"`
	Status        int16   `view:"man" json:"status"`
	Level         int16   `view:"man" json:"level"`
	Channel       string  `view:"man" json:"channel"`
	Platform      string  `view:"man" json:"platform"`
	Version       string  `view:"man" json:"version"`
	DeviceID      string  `view:"man" json:"deviceID"`
	IP            string  `view:"man" json:"ip"`
	CreateTime    int64   `view:"*" json:"createTime"`
	UpdateTime    int64   `view:"man" json:"updateTime"`
	Profile       Profile `view:"man" json:"profile"`
}

// TableName 表名
func (u *UserInfo) TableName() string {
	return "user"
}

// UserDao 用户Dao
type UserDao struct {
	*AccountBaseDao
}

// NewUserDao 创建新的UserDao
func NewUserDao() (dao *UserDao) {
	dao = &UserDao{AccountBaseDao: NewAccountBaseDao(&UserInfo{})}
	return
}

// NewSharedUserDao 创建新的UserDao, 链接是共享的.
func NewSharedUserDao(shareDao *db.SharedDao) (dao *UserDao) {
	dao = &UserDao{AccountBaseDao: NewSharedAccountBaseDao(&UserInfo{}, shareDao)}
	return
}

// RandomUniqueInviteCode 生成一个随机的, 唯一的邀请码.
func (u *UserDao) RandomUniqueInviteCode(length int) (inviteCode string) {
	for i := 0; i < 16; i++ {
		inviteCode = utils.RandomInviteCode(length)
		userInfo := u.GetByInviteCode(inviteCode)
		if userInfo == nil { //说明验证码不存在,是唯一的.
			return
		}
	}

	log.Errorf("未能成功生成邀请码")
	inviteCode = ""
	return
}

// GetByTel 按手机号,用户类型查询
func (u *UserDao) GetByTel(tel string, userType int32) (user *UserInfo) {
	obj := u.Get("tel=? AND user_type = ?", tel, userType)
	if obj != nil {
		user = obj.(*UserInfo)
	}
	return
}

// GetByUsername 按用户名查询.
func (u *UserDao) GetByUsername(username string, userType int32) (user *UserInfo) {
	obj := u.Get("username=? AND user_type = ?", username, userType)

	if obj != nil {
		user = obj.(*UserInfo)
	}
	return
}

// GetByInviteCode 按邀请码查询
func (u *UserDao) GetByInviteCode(inviteCode string) (user *UserInfo) {
	obj := u.Get("invite_code=?", inviteCode)
	if obj != nil {
		user = obj.(*UserInfo)
	}
	return
}

// GetByID 按ID查询
func (u *UserDao) GetByID(id int64) (user *UserInfo) {
	obj := u.Get("id=?", id)
	if obj != nil {
		user = obj.(*UserInfo)
	}
	return
}

// MustGetByID 按ID查询,不存在会panic
func (u *UserDao) MustGetByID(id int64) (user *UserInfo) {
	obj := u.MustGet("id=?", id)
	if obj != nil {
		user = obj.(*UserInfo)
	}
	return
}

// MustGetByUID 使用UID查询.
func (u *UserDao) MustGetByUID(uid string) (user *UserInfo) {
	obj := u.MustGet("uid=?", uid)
	if obj != nil {
		user = obj.(*UserInfo)
	}
	return
}

// SetExtraInfo 设置邀请码及uid.
func (u *UserDao) SetExtraInfo(id int64, uid string, inviteCode string) {
	now := utils.Now()
	update := map[string]interface{}{"uid": uid, "invite_code": inviteCode, "update_time": now}
	log.Infof("SetExtraInfo(id: %d, uid: %s, invite_code: %s)...", id, uid, inviteCode)
	u.UpdateBy(update, "id = ?", id)
}

// SetRegInviteCode 设置注册邀请码(被邀请的)
func (u *UserDao) SetRegInviteCode(id int64, regInviteCode string) {
	now := utils.Now()
	update := map[string]interface{}{"reg_invite_code": regInviteCode, "update_time": now}
	log.Infof("SetRegInviteCode(id: %d, reg_invite_code: %s)...", id, regInviteCode)
	u.UpdateBy(update, "id = ?", id)
}

// SetPassword 设置密码.
func (u *UserDao) SetPassword(id int64, password string) {
	now := utils.Now()
	update := map[string]interface{}{"password": password, "update_time": now}
	log.Infof("SetPassword(id: %d, password: %s)...", id, password)
	u.UpdateBy(update, "id = ?", id)
}

// SetUsername 设置密码.
func (u *UserDao) SetUsername(id int64, username string) {
	now := utils.Now()
	update := map[string]interface{}{"username": username, "update_time": now}
	log.Infof("SetUsername(id: %d, username: %s)...", id, username)
	u.UpdateBy(update, "id = ?", id)
}

// SetTel 设置手机号
func (u *UserDao) SetTel(id int64, tel string) {
	now := utils.Now()
	update := map[string]interface{}{"tel": tel, "update_time": now}
	log.Infof("SetTel(id: %d, tel: %s)...", id, tel)
	u.UpdateBy(update, "id = ?", id)
}

// SetProfile 设置profile
func (u *UserDao) SetProfile(id int64, profile string) {
	now := utils.Now()
	update := map[string]interface{}{"profile": profile, "update_time": now}
	u.UpdateBy(update, "id = ?", id)
}

// Create 创建用户
func (u *UserDao) Create(userInfo *UserInfo, inviteCodeLen int) {
	u.MustBegin()

	defer func() {
		u.Rollback()
	}()

	u.MustInsert(userInfo)
	userInfo.UID = utils.Uint64ToUID(userInfo.ID)
	if userInfo.InviteCode == "" && inviteCodeLen > 0 {
		userInfo.InviteCode = u.RandomUniqueInviteCode(inviteCodeLen)
	}
	u.SetExtraInfo(userInfo.ID, userInfo.UID, userInfo.InviteCode)

	u.MustCommit()
}
