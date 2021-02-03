package configs

import (
	"fmt"
	"io/ioutil"
	"time"

	"open-account/pkg/baselib/cache"
	"open-account/pkg/baselib/log"

	"gopkg.in/yaml.v2"
)

// CORS 跨域请求配置参数
type CORS struct {
	Enable           bool   `toml:"enable"`
	AllowOrigins     string `toml:"allow_origins"`
	AllowMethods     string `toml:"allow_methods"`
	AllowHeaders     string `toml:"allow_headers"`
	AllowCredentials bool   `toml:"allow_credentials"`
	MaxAge           int    `toml:"max_age"`
}

// DatabaseConfig 数据库配置.
type DatabaseConfig struct {
	Dialect      string `yaml:"dialect"`
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	Database     string `yaml:"database"`
	User         string `yaml:"user"`
	Password     string `yaml:"password"`
	MaxOpenConns int    `yaml:"max_open_conns"` //最大链接数.
	MaxIdleConns int    `yaml:"max_idle_conns"` //最大闲置链接数
	Debug        bool
}

// ToURL 获取dns链接
func (p *DatabaseConfig) ToURL() (url string) {
	if p.Dialect == "postgres" {
		url = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			p.Host, p.Port, p.User, p.Password, p.Database)
	} else {
		url = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			p.User, p.Password, p.Host, p.Port, p.Database)
	}
	return
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Listen            string
	LogLevel          string `yaml:"log_level"`
	LogFileName       string `yaml:"log_filename"`
	CheckSign         bool   `yaml:"check_sign"`
	Debug             bool   `yaml:"debug"`
	DisableStacktrace bool   `yaml:"disable_stacktrace"`
	URLSign           bool   `yaml:"url_sign"`
	CheckCaptcha      bool   `yaml:"check_captcha"`
	ServerEnv         string `yaml:"server_env"`

	// 客户端测试账号,(不会真正发送验证码) key为手机号, value为验证码.
	TestAccounts map[string]string `yaml:"test_accounts"`
	// 给单元测试使用的超级验证码, 只在Debug=true时有效.
	SuperCodeForTest string `yaml:"super_code_for_test"`
	// 访问`/v1/man/account/sms/get/code`接口使用的key.
	SuperKeyForTest string `yaml:"super_key_for_test"`
	// 超级签名, 仅用于测试, 只在Debug=true时有效
	SuperSignKey string `yaml:"super_sign_key"`
	// 访问后台接口(以 `/v1/man/` 开头的)需要使用该token.
	AdminToken string `yaml:"admin_token"`

	InviteCodeSettingPeriod time.Duration `yaml:"invite_code_setting_period"`

	ReqDebug     bool          `yaml:"req_debug"`      //是否开启请求日志
	ReqDebugHost string        `yaml:"req_debug_host"` //host值, 默认为: https://127.0.0.1:2021
	ReqDebugDir  string        `yaml:"req_debug_dir"`  //目录值, 默认为: /data/logs/req_debug/
	FdExpiretime time.Duration `yaml:"fd_expiretime"`  //文件句柄过期时间,默认为10分钟.

	InviteCodeLength int `yaml:"invite_code_length"` // 生成的邀请码长度

	AccountDB  *DatabaseConfig    `yaml:"account_database"` //mysql 链接配置.
	TokenRedis *cache.RedisConfig `yaml:"token_redis"`      // TOKEN
	SmsRedis   *cache.RedisConfig `yaml:"sms_redis"`

	CORS CORS `yaml:"cors"`

	// appID及对应的sign key
	AppKeys     map[string]string `yaml:"app_keys"`
	SignHeaders []string          `yaml:"sign_headers"` //需要参与签名的请求头列表

	CustomHeaderPrefix string `yaml:"custom_header_prefix"` //自定义请求头的前缀(默认为X-OA-)
}

// Config 全局配置
var Config *ServerConfig

// GetSignKey 获取应用的signKey
func (a *ServerConfig) GetSignKey(appID string) string {
	return a.AppKeys[appID]
}

// LoadConfig 加载解析配置.
func LoadConfig(configFilename string) (config *ServerConfig, err error) {
	Config = &ServerConfig{
		Listen:       ":2021",
		LogLevel:     "debug",
		LogFileName:  "./logs/open-account.log",
		CheckSign:    true,
		Debug:        false,
		URLSign:      false,
		CheckCaptcha: false,
		ServerEnv:    "dev",
		ReqDebug:     true,
		ReqDebugHost: "http://127.0.0.1:2021",
		ReqDebugDir:  "/data/logs/req_debug/",
		FdExpiretime: time.Minute * 10,

		InviteCodeSettingPeriod: time.Hour * 24 * 7, // 默认7天内可设置邀请码.
		CustomHeaderPrefix:      "X-OA-",

		AccountDB: &DatabaseConfig{
			Dialect:      "mysql",
			Host:         "127.0.0.1",
			Port:         3306,
			Database:     "openaccount",
			User:         "openaccount",
			Password:     "123456",
			Debug:        true,
			MaxOpenConns: 50,
			MaxIdleConns: 20,
		},
		TokenRedis: &cache.RedisConfig{
			CacheName: "token",
			Addr:      "127.0.0.1:6379",
			Password:  "",
			DBIndex:   1,
			Exptime:   time.Hour * 24 * 30,
		},
		SmsRedis: &cache.RedisConfig{
			CacheName: "sms",
			Addr:      "127.0.0.1:6379",
			Password:  "",
			DBIndex:   2,
			Exptime:   time.Second * 180,
		},
	}
	config = Config

	config.AppKeys = make(map[string]string)

	var buf []byte
	buf, err = ioutil.ReadFile(configFilename)
	if err != nil {
		log.Fatalf("LoadConfig(%s) failed! err: %v", configFilename, err)
		return
	}
	err = yaml.Unmarshal(buf, Config)
	if err != nil {
		log.Fatalf("Unmarshal yaml config failed! err: %v", configFilename, err)
		return
	}

	return
}
