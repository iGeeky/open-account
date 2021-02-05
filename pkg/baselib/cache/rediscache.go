package cache

import (
	"fmt"
	"github.com/iGeeky/open-account/pkg/baselib/log"
	"strings"
	"time"

	"gopkg.in/redis.v5"
)

var (
	// ErrNotExist 数据不存在
	ErrNotExist error
)

func init() {
	ErrNotExist = redis.Nil
}

// RedisConfig Redis配置.
type RedisConfig struct {
	CacheName string `yaml:"cache_name"`
	Addr      string // host:port
	Password  string
	DBIndex   int `yaml:"db_index"`
	Exptime   time.Duration
}

// DebugStr 配置概要
func (r *RedisConfig) DebugStr() (debugStr string) {
	debugStr = fmt.Sprintf("{CacheName: '%s', Addr: '%s', Index: %d, Exptime: '%v'}", r.CacheName, r.Addr, r.DBIndex, r.Exptime)
	return
}

// RedisCache Redis缓存
type RedisCache struct {
	cacheName string
	Cfg       *RedisConfig
	client    *redis.Client
}

// NewRedisCache 使用指定配置,创建一个Redis缓存.
func NewRedisCache(cfg *RedisConfig) (cache *RedisCache, err error) {
	cache = &RedisCache{}

	cache.cacheName = cfg.CacheName
	if cache.cacheName == "" {
		cache.cacheName = fmt.Sprintf("redis_%s", cfg.Addr)
	}
	if cfg.Addr == "" {
		cfg.Addr = "127.0.0.1:6379"
	}
	cache.Cfg = cfg

	cache.client = redis.NewClient(&redis.Options{Addr: cfg.Addr, Password: cfg.Password, DB: cfg.DBIndex})
	pong, err := cache.client.Ping().Result()
	if err != nil {
		log.Errorf("NewRedisCache(%v,%v) failed! pong:%v err:%v", cfg.Addr, cfg.DBIndex, pong, err)
		cache = nil
		return
	}

	return
}

// Expire 设置过期时间.
func (r *RedisCache) Expire(key string, exptime time.Duration) (err error) {
	err = r.client.Expire(key, exptime).Err()
	if err != nil {
		log.Errorf("redis.Expire(%s, %v) failed! err:%v", key, exptime, err)
		err = nil
	}

	return
}

// Keys 输出指定pattern的keys
func (r *RedisCache) Keys(pattern string) (keys []string, err error) {
	result := r.client.Keys(pattern)
	keys, err = result.Result()
	return
}

// GetB 获取字节数组数据, 如果key为 xxx->field, 将从hash中获取数据.
func (r *RedisCache) GetB(key string) (val []byte, err error) {
	keys := strings.SplitN(key, "->", 2)
	if len(keys) == 2 {
		val, err = r.HGetB(keys[0], keys[1])
	} else {
		val, err = r.client.Get(key).Bytes()
	}
	return
}

// Get 获取字符串数据, 如果key为 xxx->field, 将从hash中获取数据.
func (r *RedisCache) Get(key string) (val string, err error) {
	b, err := r.GetB(key)
	if err == nil {
		val = string(b)
	}
	return
}

// IncrBy 累加指定的key
func (r *RedisCache) IncrBy(key string, value int64) (i int64, err error) {
	i, err = r.client.IncrBy(key, value).Result()
	return
}

// Set 设置值, 如果key为 xxx->field, 将数据存储到hash中.
func (r *RedisCache) Set(key string, value interface{}, exptime time.Duration) (err error) {
	keys := strings.SplitN(key, "->", 2)
	if len(keys) == 2 {
		err = r.HSet(keys[0], keys[1], value, exptime)
	} else {
		err = r.client.Set(key, value, exptime).Err()
	}
	return
}

// HGet 获取Hash的指定字段的值.
func (r *RedisCache) HGet(key, field string) (val string, err error) {
	val, err = r.client.HGet(key, field).Result()
	return
}

// HGetI 获取Hash的指定字段的值(int64)
func (r *RedisCache) HGetI(key, field string) (val int64, err error) {
	val, err = r.client.HGet(key, field).Int64()
	return
}

// HGetB 获取Hash的指定字段的值([]byte)
func (r *RedisCache) HGetB(key, field string) (val []byte, err error) {
	val, err = r.client.HGet(key, field).Bytes()
	return
}

// HSet 设置Hash的指定字段的值
func (r *RedisCache) HSet(key, field string, value interface{}, exptime time.Duration) (err error) {
	err = r.client.HSet(key, field, value).Err()
	if err == nil && exptime > 0 {
		err = r.Expire(key, exptime)
	}
	return
}

// HIncrBy 对Hash指定字段进行累加
func (r *RedisCache) HIncrBy(key, field string, incr int64) (n int64, err error) {
	n, err = r.client.HIncrBy(key, field, incr).Result()
	return
}

// Del 删除指定的key
func (r *RedisCache) Del(keys ...string) (n int64, err error) {
	n, err = r.client.Del(keys...).Result()
	return
}
