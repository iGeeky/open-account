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

// MysqlConfig Mysql配置.
type MysqlConfig struct {
	Host         string
	Port         int
	DBName       string `yaml:"db_name"`
	User         string
	Password     string
	MaxOpenConns int `yaml:"max_open_conns"` //最大链接数.
	MaxIdleConns int `yaml:"max_idle_conns"` //最大闲置链接数
	Debug        bool
}

func (p *MysqlConfig) ToURL() (url string) {
	url = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		p.Host, p.Port, p.User, p.Password, p.DBName)
	return
}

func (p *MysqlConfig) ToSecretURL() (url string) {
	password := p.Password
	url = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		p.Host, p.Port, p.User, password, p.DBName)
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

	ServerEnv string `yaml:"server_env"`
	// 客户端测试账号,(不会真正发送验证码) key为手机号, value为验证码.
	TestAccounts map[string]string `yaml:"test_accounts"`
	// 给单元测试使用的超级验证码, 只在Debug=true时有效果.
	SuperCodeForTest string `yaml:"super_code_for_test"`

	ReqDebug     bool          `yaml:"req_debug"`      //是否开启请求日志
	ReqDebugHost string        `yaml:"req_debug_host"` //host值, 默认为: https://127.0.0.1:2021
	ReqDebugDir  string        `yaml:"req_debug_dir"`  //目录值, 默认为: /data/logs/req_debug/
	FdExpiretime time.Duration `yaml:"fd_expiretime"`  //文件句柄过期时间,默认为10分钟.

	MySQL      *MysqlConfig       `yaml:"mysql"`       //mysql 链接配置.
	TokenRedis *cache.RedisConfig `yaml:"token_redis"` // TOKEN
	SmsRedis   *cache.RedisConfig `yaml:"sms_redis"`

	CORS CORS `yaml:"cors"`

	// appID及对应的sign key
	AppKeys map[string]string `yaml:"app_keys"`
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
		ServerEnv:    "dev",
		ReqDebug:     true,
		ReqDebugHost: "http://127.0.0.1:2021",
		ReqDebugDir:  "/data/logs/req_debug/",
		FdExpiretime: time.Minute * 10,

		MySQL: &MysqlConfig{
			Host:         "127.0.0.1",
			Port:         3306,
			DBName:       "open_account",
			User:         "open_account",
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
