package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// go get -u go.uber.org/zap

// Logger 全局日志对象
var Logger *zap.Logger

func init() {
	// 配置日志输出到文件
	zapConfig := zap.NewProductionConfig()
	zapConfig.OutputPaths = []string{"log.log", "stdout"} // 将日志输出到文件 和 标准输出
	zapConfig.Encoding = "console"                        // 设置日志格 json console
	zapConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel) // 设置日志级别
	zapConfig.EncoderConfig = zapcore.EncoderConfig{      // 创建Encoder配置
		MessageKey:   "message",
		LevelKey:     "level",
		EncodeLevel:  zapcore.LowercaseLevelEncoder,
		TimeKey:      "time",
		EncodeTime:   zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05"),
		CallerKey:    "caller",
		EncodeCaller: zapcore.ShortCallerEncoder,
	}
	//zapConfig.Sampling = nil

	// 创建Logger对象
	var err error
	Logger, err = zapConfig.Build()
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	// 在应用程序退出时调用以确保所有日志消息都被写入文件
	defer func(Logger *zap.Logger) {
		_ = Logger.Sync()
	}(Logger)
}
