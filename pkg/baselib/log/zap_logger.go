package log

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger
var sugarLogger *zap.SugaredLogger

// InitLogger 初始化Logger
func InitLogger(logLevel, logFilename string, development, disableStacktrace bool) *zap.SugaredLogger {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder, // 小写编码器
		EncodeTime:     zapcore.ISO8601TimeEncoder,    // ISO8601 UTC 时间格式
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.FullCallerEncoder, // 全路径编码器
	}

	// 设置日志级别
	level := zap.NewAtomicLevel()
	level.UnmarshalText([]byte(logLevel))

	OutputPaths := []string{"stdout"}
	ErrorOutputPaths := []string{"stderr"}
	if logFilename != "" {
		OutputPaths = append(OutputPaths, logFilename)
		ErrorOutputPaths = append(ErrorOutputPaths, logFilename)
	}

	config := zap.Config{
		Level:             level,             // 日志级别
		Development:       development,       // 开发模式，堆栈跟踪
		DisableStacktrace: disableStacktrace, // 关闭堆栈追踪
		Encoding:          "console",         // 输出格式 console 或 json
		EncoderConfig:     encoderConfig,     // 编码器配置
		// InitialFields:    map[string]interface{}{"service": "github.com/iGeeky/open-account"}, // 初始化字段，如：添加一个服务器名称
		OutputPaths:      OutputPaths, // 输出到指定文件 stdout（标准输出，正常颜色） stderr（错误输出，红色）
		ErrorOutputPaths: ErrorOutputPaths,
	}

	var err error
	// 构建日志
	logger, err = config.Build()
	if err != nil {
		panic(fmt.Sprintf("log init failed: %v", err))
	}

	tmpLogger, err := config.Build(zap.AddCallerSkip(1))
	if err != nil {
		panic(fmt.Sprintf("log init failed: %v", err))
	}

	sugarLogger = tmpLogger.Sugar()

	return sugarLogger
}

func init() {
	logger, _ = zap.NewDevelopment()
	tmpLogger, _ := zap.NewDevelopment(zap.AddCallerSkip(1))
	sugarLogger = tmpLogger.Sugar()
}

func Logger() *zap.Logger {
	return logger
}

func Debug(args ...interface{}) {
	sugarLogger.Debug(args...)
}

func Debugf(template string, args ...interface{}) {
	sugarLogger.Debugf(template, args...)
}

func Info(args ...interface{}) {
	sugarLogger.Info(args...)
}

func Infof(template string, args ...interface{}) {
	sugarLogger.Infof(template, args...)
}

func Warn(args ...interface{}) {
	sugarLogger.Warn(args...)
}

func Warnf(template string, args ...interface{}) {
	sugarLogger.Warnf(template, args...)
}

func Error(args ...interface{}) {
	sugarLogger.Error(args...)
}

func Errorf(template string, args ...interface{}) {
	sugarLogger.Errorf(template, args...)
}

func DPanic(args ...interface{}) {
	sugarLogger.DPanic(args...)
}

func DPanicf(template string, args ...interface{}) {
	sugarLogger.DPanicf(template, args...)
}

func Panic(args ...interface{}) {
	sugarLogger.Panic(args...)
}

func Panicf(template string, args ...interface{}) {
	sugarLogger.Panicf(template, args...)
}

func Fatal(args ...interface{}) {
	sugarLogger.Fatal(args...)
}

func Fatalf(template string, args ...interface{}) {
	sugarLogger.Fatalf(template, args...)
}
