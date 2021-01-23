package db

import (
	"database/sql"
	"fmt"
	stdlog "log"
	"open-account/pkg/baselib/errors"
	"os"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"

	"open-account/pkg/baselib/log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

// ITableName 表名获取接口
type ITableName interface {
	TableName() string
}

// Now 当前unix时间
func Now() (now int64) {
	now = time.Now().Unix()
	return
}

var (
	gMaxOpenConns = 1024
	gMaxIdleConns = 128
)

// InitDBConfig 初始化数据库配置
func InitDBConfig(maxOpenConns, maxIdleConns int) {
	gMaxOpenConns = maxOpenConns
	gMaxIdleConns = maxIdleConns
}

func uriDebug(dbURI string) string {
	regexPassword := regexp.MustCompile(`password=[0-9a-zA-Z]*`)
	return regexPassword.ReplaceAllString(dbURI, "password=******")
}

func initDBInternal(dialect, dbURI string, sqlDebug bool) (db *gorm.DB) {
	var err error
	db, err = gorm.Open(dialect, dbURI)
	if err != nil {
		log.Panicf("open database(%s) failed! err: %v", uriDebug(dbURI), err)
	} else {
		log.Infof("open database(%s) ok!", uriDebug(dbURI))
	}
	// DB.SetConnMaxLifetime(d)
	db.DB().SetMaxIdleConns(gMaxIdleConns)
	db.DB().SetMaxOpenConns(gMaxOpenConns)
	db.LogMode(sqlDebug)
	log.Infof("gMaxOpenConns(%d), gMaxIdleConns(%d), sqlDebug: %v", gMaxOpenConns, gMaxIdleConns, sqlDebug)

	if sqlDebug {
		db.SetLogger(stdlog.New(os.Stdout, "api:", stdlog.Lmicroseconds))
	}

	return
}

var dbMap sync.Map

func initDB(dialect, dbURL string, sqlDebug bool) (db *gorm.DB) {
	var ok bool
	value, ok := dbMap.Load(dbURL)
	if ok {
		db = value.(*gorm.DB)
		return
	}

	db = initDBInternal(dialect, dbURL, sqlDebug)

	dbMap.Store(dbURL, db)
	// log.Infof("open database(%s) ok!", uriDebug(dbURL))
	return
}

// SharedDao 共享数据库链接对象的Dao
type SharedDao struct {
	dialect  string
	dbURL    string
	sqlDebug bool
	db       *gorm.DB
	tx       *gorm.DB
	inTx     bool //是否处于事务中.
}

func newSharedDao(dialect, dbURL string, sqlDebug bool) (dao *SharedDao) {
	dao = &SharedDao{dialect: dialect, dbURL: dbURL, sqlDebug: sqlDebug, db: initDB(dialect, dbURL, sqlDebug)}
	return
}

// BaseDao 基础Dao
// BaseDao 中Must开头的函数, 如果出错会抛出Panic, 错误信息为: baselib.errors.ApiError{Reason: 'ERR_SERVER_ERROR', Errmsg: '错误的具体信息'}
type BaseDao struct {
	dao   *SharedDao
	model interface{}
}

// TableName 获取表名
func (p *BaseDao) TableName() (tableName string) {
	tableName = "object"
	if iTableName, ok := p.model.(ITableName); ok {
		tableName = iTableName.TableName()
	}
	return
}

// NewBaseDao 创建新的BaseDao, model指定对应的模型类
func NewBaseDao(dialect, dbURL string, sqlDebug bool, model interface{}) (dao *BaseDao) {
	dao = &BaseDao{model: model, dao: newSharedDao(dialect, dbURL, sqlDebug)}
	return
}

func NewSharedBaseDao(shared *SharedDao, model interface{}) (dao *BaseDao) {
	dao = &BaseDao{model: model, dao: shared}
	return
}

func (p *BaseDao) GetShared() (shared *SharedDao) {
	shared = p.dao
	return
}

func (p *BaseDao) DB() (db *gorm.DB) {
	if p.dao.inTx {
		return p.dao.tx
	} else {
		return p.dao.db
	}
}

// Begin 开启事务
func (p *BaseDao) Begin() (err error) {
	if p.dao.inTx {
		log.Errorf("[%s] Begin transaction failed! a transaction is begin", p.TableName())
		return
	}
	p.dao.tx = p.dao.db.Begin()
	err = p.dao.tx.Error
	if err != nil {
		p.dao.tx = nil
		log.Errorf("[%s] Begin transaction failed! err: %v", p.TableName(), err)
		return
	}
	p.dao.inTx = true
	return
}

// Commit 提交事务
func (p *BaseDao) Commit() (err error) {
	if !p.dao.inTx {
		log.Errorf("[%s] Commit transaction failed! no transaction is begin", p.TableName())
		return
	}
	err = p.dao.tx.Commit().Error
	if err != nil {
		log.Errorf("[%s] Commit transaction failed! err: %v", p.TableName(), err)
		return
	}
	p.dao.inTx = false
	p.dao.tx = nil
	return
}

// Rollback 回滚事务
func (p *BaseDao) Rollback() (err error) {
	if !p.dao.inTx {
		log.Debugf("[%s] rollback transaction failed! no transaction is begin", p.TableName())
		return
	}
	err = p.dao.tx.RollbackUnlessCommitted().Error
	if err != nil {
		log.Errorf("[%s] Rollback transaction failed! err: %v", p.TableName(), err)
		return
	}
	p.dao.inTx = false
	p.dao.tx = nil
	return
}

// MustBegin 开始事务, 如果出错会导致panic.
func (p *BaseDao) MustBegin() {
	errors.Assert(p.Begin())
}

func (p *BaseDao) MustCommit() {
	errors.Assert(p.Commit())
}

func (p *BaseDao) MustRollback() {
	errors.Assert(p.Rollback())
}

func (p *BaseDao) Insert(value interface{}) (err error) {
	db := p.DB().Create(value)
	if db.Error != nil {
		log.Errorf("[%s] Insert(%+v) failed! err: %v", p.TableName(), value, db.Error)
		err = db.Error
	}
	return
}

func (p *BaseDao) MustInsert(value interface{}) {
	errors.Assert(p.Insert(value))
}

func (p *BaseDao) Save(value interface{}) (err error) {
	db := p.DB().Save(value)
	if db.Error != nil {
		log.Errorf("[%s] Save(%+v) failed! err: %v", p.TableName(), value, db.Error)
		err = db.Error
	}
	return
}

func (p *BaseDao) MustSave(value interface{}) {
	errors.Assert(p.Save(value))
}

// 有设置主键时才可以使用. 唯一键不行.
func (p *BaseDao) Upsert(obj interface{}) (operator string, effectRows int64) {
	operator = "save"
	db := p.DB().Create(obj)
	err := db.Error
	if err != nil && strings.Index(err.Error(), "duplicate key value violates unique constraint") > 0 {
		operator = "update"
		dbUpdate := p.DB().Model(p.model).Update(obj)
		err = dbUpdate.Error
		effectRows = dbUpdate.RowsAffected
	} else {
		effectRows = db.RowsAffected
	}
	if err != nil {
		log.Errorf("[%s] Upsert(%+v) failed! err: %v", p.TableName(), obj, err)
		errors.Assert(err)
	}

	return
}

// UpdateBy 根据查询,更新对象, 结构体中为空("")的字段不会被重置为空
func (p *BaseDao) UpdateBy(update interface{}, query interface{}, args ...interface{}) (effectRows int64) {
	db := p.DB().Model(p.model).Where(query, args...).Updates(update, false)
	err := db.Error
	if err != nil {
		log.Errorf("[%s] Update(update=%+v, query=%+v) failed! err: %v", p.TableName(), update, query, err)
		errors.Assert(err)
	} else {
		effectRows = db.RowsAffected
	}
	return
}

// Delete 根据查询及参数删除
func (p *BaseDao) Delete(query interface{}, args ...interface{}) (effectRows int64) {
	db := p.DB().Model(p.model).Where(query, args...).Delete(p.model)
	if err := db.Error; err != nil {
		log.Errorf("[%s] Delete(query=%+v) failed! err: %v", p.TableName(), query, err)
		errors.Assert(err)
	} else {
		effectRows = db.RowsAffected
	}
	return
}

// DeleteBy 按指定的字段及值删除
func (p *BaseDao) DeleteBy(fieldName string, value interface{}) (effectRows int64) {
	effectRows = p.Delete(fieldName+" = ? ", value)
	return
}

// DeleteByID 根据id删除.
func (p *BaseDao) DeleteByID(id interface{}) (effectRows int64) {
	effectRows = p.Delete("id = ? ", id)
	return
}

// Find 简单查询
func (p *BaseDao) Find(query interface{}, args ...interface{}) (result interface{}, err error) {
	eleType := reflect.ValueOf(p.model).Elem().Type()
	result = reflect.New(reflect.SliceOf(eleType)).Interface()
	err = p.DB().Where(query, args...).Find(result).Error
	if err != nil {
		result = nil
	}

	return
}

//PagingFind 支持分页的查询, 返回值result 类型为: *[]Object
func (p *BaseDao) PagingFind(orderBy interface{}, page int, pageSize int, query interface{}, args ...interface{}) (result interface{}, err error) {
	eleType := reflect.ValueOf(p.model).Elem().Type()
	result = reflect.New(reflect.SliceOf(eleType)).Interface()
	if page > 0 {
		page = page - 1
	}
	if pageSize < 1 {
		pageSize = 1
	}
	offset := page * pageSize
	limit := pageSize
	db := p.DB().Model(p.model).Where(query, args...)
	if orderBy != nil {
		db = db.Order(orderBy)
	}

	err = db.Offset(offset).Limit(limit).Find(result).Error
	if err != nil {
		result = nil
	}

	return
}

// MustPagingFind 功能同PagingFind, 如果出错会导致panic.
func (p *BaseDao) MustPagingFind(orderBy interface{}, page int, pageSize int, query interface{}, args ...interface{}) (result interface{}) {
	var err error
	result, err = p.PagingFind(orderBy, page, pageSize, query, args...)
	if err != nil {
		log.Errorf("[%s] FindEx(query=%+v, args=%+v, orderBy=%+v)", p.TableName(), query, args, orderBy)
		errors.Assert(err)
	}
	return
}

// RawFindModel 使用原生SQL查询(指定结果Model)
func (p *BaseDao) RawFindModel(model interface{}, rawQuery string, args ...interface{}) (result interface{}, err error) {
	var rows *sql.Rows
	rows, err = p.DB().Raw(rawQuery, args...).Rows()
	if err != nil {
		return
	}
	defer rows.Close()

	eleType := reflect.ValueOf(model).Elem().Type()
	result = reflect.New(reflect.SliceOf(eleType)).Interface()
	resultValue := reflect.Indirect(reflect.ValueOf(result))

	for rows.Next() {
		var obj = reflect.New(eleType).Interface()
		err = p.DB().ScanRows(rows, obj)
		if err != nil {
			return
		}
		resultValue = reflect.Append(resultValue, reflect.ValueOf(obj).Elem())
	}
	result = resultValue.Interface()

	return
}

// MustRawFindModel 功能同RawFindModel, 如果出错会导致panic.
func (p *BaseDao) MustRawFindModel(model interface{}, rawQuery string, args ...interface{}) (result interface{}) {
	var err error
	result, err = p.RawFindModel(model, rawQuery, args...)
	errors.Assert(err)
	return
}

// RawFind 使用原生SQL查询(使用当前dao的模型)
func (p *BaseDao) RawFind(rawQuery string, args ...interface{}) (result interface{}, err error) {
	result, err = p.RawFindModel(p.model, rawQuery, args...)
	if err != nil {
		log.Errorf("[%s] RawFind(query: '%v', args: %v) failed! err: %v", p.TableName(), rawQuery, args, err)
	}
	return
}

// MustRawFind 功能同RawFind, 如果出错会导致panic.
func (p *BaseDao) MustRawFind(rawQuery string, args ...interface{}) (result interface{}) {
	var err error
	result, err = p.RawFind(rawQuery, args...)
	errors.Assert(err)
	return
}

// GetSimple 简单查询
func (p *BaseDao) GetSimple(query interface{}, args ...interface{}) (result interface{}, err error) {
	eleType := reflect.ValueOf(p.model).Elem().Type()
	result = reflect.New(eleType).Interface()
	err = p.DB().Where(query, args...).First(result).Error
	if err == gorm.ErrRecordNotFound {
		result = nil
		err = nil
	}
	if err != nil {
		log.Errorf("[%s] MustGet(query: '%v', args: %v) failed! err: %v", p.TableName(), query, args, err)
	}
	return
}

// Get 功能同GetSimple, 如果出错会导致panic.
func (p *BaseDao) Get(query interface{}, args ...interface{}) (result interface{}) {
	var err error
	result, err = p.GetSimple(query, args...)
	errors.Assert(err)
	return
}

// MustGet 功能同Get, 如果未查询到对象, 会panic(ErrObjectNotFound)
func (p *BaseDao) MustGet(query interface{}, args ...interface{}) (result interface{}) {
	result = p.Get(query, args...)
	if result == nil {
		log.Infof("%s.MustGet(where: '%+v', args: %+v) not found", p.TableName(), query, args)
		if p.dao.sqlDebug {
			panic(errors.NewError(errors.ErrObjectNotFound, fmt.Sprintf("%s.MustGet(where: '%+v', args: %+v) not found", p.TableName(), query, args)))
		} else {
			panic(errors.NewError(errors.ErrObjectNotFound, "object not found"))
		}
	}
	return
}

// MustRawGet 使用原始查询,查询单个对象
func (p *BaseDao) MustRawGet(model interface{}, rawQuery string, args ...interface{}) (result interface{}) {
	rows, err := p.DB().Raw(rawQuery, args...).Rows()
	eleType := reflect.ValueOf(model).Elem().Type()
	if err == nil {
		defer rows.Close()
		if rows.Next() {
			result = reflect.New(eleType).Interface()
			err = p.DB().ScanRows(rows, result)
		} else {
			result = nil
		}
	}
	if err != nil {
		log.Errorf("%v.MustRawGet(query: '%v', args: %v) failed! err: %v", eleType, rawQuery, args, err)
		errors.Assert(err)
	}
	return
}

// Exec 执行Update更新语句.
func (p *BaseDao) Exec(sql string, args ...interface{}) (affectedRows int64, err error) {
	db := p.DB().Exec(sql, args...)
	affectedRows = db.RowsAffected
	err = db.Error
	return
}

// MustExec 功能周Exec, 如果出错会导致panic.
func (p *BaseDao) MustExec(sql string, args ...interface{}) (affectedRows int64) {
	affectedRows, err := p.Exec(sql, args...)
	if err != nil {
		log.Errorf("[%s] MustExec(sql: '%v', args: %v) failed! err: %v", p.TableName(), sql, args, err)
		errors.Assert(err)
	}
	return
}

// QueryInt 执行查询,返回Int结果
func (p *BaseDao) QueryInt(sql string, args ...interface{}) (value int64, err error) {
	row := p.DB().Raw(sql, args...).Row()
	err = row.Scan(&value)
	if err != nil {
		log.Errorf("[%s] QueryInt(sql: '%v', args: %v) failed! err: %v", p.TableName(), sql, args, err)
	}
	return
}

// MustQueryInt 功能同QueryInt, 如果出错会导致panic.
func (p *BaseDao) MustQueryInt(sql string, args ...interface{}) (value int64) {
	value, err := p.QueryInt(sql, args...)
	errors.Assert(err)
	return
}

// MustCount 执行Count查询. 如果出错会导致panic.
func (p *BaseDao) MustCount(query string, args ...interface{}) (count int64) {
	errors.Assert(p.DB().Model(p.model).Where(query, args...).Count(&count).Error)
	return
}
