package dao

import (
	// "github.com/lib/pq"
	"github.com/iGeeky/open-account/configs"
	"github.com/iGeeky/open-account/pkg/baselib/db"
)

// AccountBaseDao 账号系统,基础Dao.
type AccountBaseDao struct {
	*db.BaseDao
}

// NewAccountBaseDao model是dao关联的模型
func NewAccountBaseDao(model interface{}) (dao *AccountBaseDao) {
	dao = &AccountBaseDao{BaseDao: db.NewBaseDao(configs.Config.AccountDB.Dialect, configs.Config.AccountDB.ToURL(), configs.Config.AccountDB.Debug, model)}
	return
}

// NewSharedAccountBaseDao 创建共享链接的dao.
func NewSharedAccountBaseDao(model interface{}, shared *db.SharedDao) (dao *AccountBaseDao) {
	dao = &AccountBaseDao{BaseDao: db.NewSharedBaseDao(shared, model)}
	return
}
