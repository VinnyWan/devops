package logger

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

// Init 自定义 zap 日志器（JSON + 东八区时间 + 自定义字段名）
func Init() {
	encoder := getEncoder()
	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout), // 这里输出到控制台，如需输出到文件可以换成 file writer
		zap.NewAtomicLevelAt(zap.InfoLevel),
	)

	// 添加 caller、stacktrace 等配置
	Log = zap.New(core,
		zap.AddCaller(),                   // 打印调用方信息 caller
		zap.AddCallerSkip(1),              // 跳过一层封装，如有需要可调整
		zap.AddStacktrace(zap.ErrorLevel), // error 级别及以上打印堆栈
	)
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
