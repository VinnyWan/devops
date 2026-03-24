package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Log *zap.Logger

// LogConfig 日志配置参数
type LogConfig struct {
	Level            string // debug, info, warn, error
	Output           string // console, file, both
	FilePath         string // 日志文件路径
	EnableCaller     bool
	EnableStacktrace bool
}

// Init 使用默认配置初始化（配置加载前的临时 logger）
func Init() error {
	return InitWithConfig(&LogConfig{
		Level:            "info",
		Output:           "console",
		EnableCaller:     true,
		EnableStacktrace: true,
	})
}

// InitWithConfig 根据配置初始化日志系统
func InitWithConfig(cfg *LogConfig) error {
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		level = zapcore.InfoLevel
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		MessageKey:     "msg",
		CallerKey:      "caller",
		StacktraceKey:  "stack",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var cores []zapcore.Core

	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
	jsonEncoder := zapcore.NewJSONEncoder(encoderConfig)

	switch cfg.Output {
	case "file":
		if cfg.FilePath != "" {
			fileWriter := newFileWriter(cfg.FilePath)
			cores = append(cores, zapcore.NewCore(jsonEncoder, fileWriter, level))
		}
	case "both":
		cores = append(cores, zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), level))
		if cfg.FilePath != "" {
			fileWriter := newFileWriter(cfg.FilePath)
			cores = append(cores, zapcore.NewCore(jsonEncoder, fileWriter, level))
		}
	default:
		cores = append(cores, zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), level))
	}

	if len(cores) == 0 {
		cores = append(cores, zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), level))
	}

	core := zapcore.NewTee(cores...)

	var opts []zap.Option
	if cfg.EnableCaller {
		opts = append(opts, zap.AddCaller())
	}
	if cfg.EnableStacktrace {
		opts = append(opts, zap.AddStacktrace(zapcore.ErrorLevel))
	}

	Log = zap.New(core, opts...)
	return nil
}

func newFileWriter(filePath string) zapcore.WriteSyncer {
	return zapcore.AddSync(&lumberjack.Logger{
		Filename:   filePath,
		MaxSize:    100, // MB
		MaxBackups: 5,
		MaxAge:     30, // 天
		Compress:   true,
	})
}
