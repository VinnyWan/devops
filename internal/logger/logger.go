package logger

import (
	"os"
	"path/filepath"
	"time"

	"devops/common/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

// Init 初始化日志器，根据配置输出到终端或文件
func Init() {
	logConfig := config.GetLogConfig()

	// 获取日志级别
	level := getLogLevel(logConfig.Level)

	// 获取编码器
	encoder := getEncoder()

	// 获取输出目标
	writeSyncer := getWriteSyncer(logConfig)

	// 创建 core
	core := zapcore.NewCore(
		encoder,
		writeSyncer,
		zap.NewAtomicLevelAt(level),
	)

	// 添加配置选项
	opts := []zap.Option{}
	if logConfig.EnableCaller {
		opts = append(opts, zap.AddCaller(), zap.AddCallerSkip(1))
	}
	if logConfig.EnableStacktrace {
		opts = append(opts, zap.AddStacktrace(zap.ErrorLevel))
	}

	Log = zap.New(core, opts...)
}

// 日志编码器：JSON + 自定义时间/字段名/级别格式
func getEncoder() zapcore.Encoder {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",   // 时间字段名
		LevelKey:       "level",  // 日志级别字段名
		NameKey:        "logger", // logger 名称字段名
		CallerKey:      "caller", // 调用方字段名
		MessageKey:     "msg",    // 日志消息字段名
		StacktraceKey:  "stack",  // 堆栈字段名
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder, // INFO / ERROR
		EncodeCaller:   zapcore.ShortCallerEncoder,  // main.go:25
		EncodeDuration: zapcore.SecondsDurationEncoder,
		// 关键：自定义时间编码为东八区 + 人类可读格式
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			// 转到东八区
			cst := time.FixedZone("CST", 8*3600)
			enc.AppendString(t.In(cst).Format("2006-01-02 15:04:05.000"))
		},
	}

	return zapcore.NewJSONEncoder(encoderConfig)
}

// getLogLevel 获取日志级别
func getLogLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zap.DebugLevel
	case "info":
		return zap.InfoLevel
	case "warn":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	default:
		return zap.InfoLevel
	}
}

// getWriteSyncer 获取输出目标
func getWriteSyncer(logConfig *config.Log) zapcore.WriteSyncer {
	switch logConfig.Output {
	case "console":
		// 输出到终端
		return zapcore.AddSync(os.Stdout)
	case "file":
		// 输出到文件
		return getFileWriter(logConfig.FilePath)
	case "both":
		// 同时输出到终端和文件
		return zapcore.NewMultiWriteSyncer(
			zapcore.AddSync(os.Stdout),
			getFileWriter(logConfig.FilePath),
		)
	default:
		return zapcore.AddSync(os.Stdout)
	}
}

// getFileWriter 获取文件写入器
func getFileWriter(filePath string) zapcore.WriteSyncer {
	// 确保日志目录存在
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		panic("Failed to create log directory: " + err.Error())
	}

	// 打开或创建日志文件
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic("Failed to open log file: " + err.Error())
	}

	return zapcore.AddSync(file)
}
