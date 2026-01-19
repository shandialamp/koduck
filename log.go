package koduck

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Log struct {
	access *zap.Logger
	err    *zap.Logger
	app    *zap.Logger
}

func NewLog(cfg LogConfig) *Log {
	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())

	accessCore := zapcore.NewCore(
		encoder,
		zapcore.AddSync(&lumberjack.Logger{
			Filename:   cfg.AccessFile,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
		}),
		zap.InfoLevel, // access 记录 Info ~ Debug ~ Warn
	)

	errorCore := zapcore.NewCore(
		encoder,
		zapcore.AddSync(&lumberjack.Logger{
			Filename:   cfg.ErrorFile,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
		}),
		zap.ErrorLevel,
	)

	// ========== Framework Logger ==========
	appCore := zapcore.NewCore(
		encoder,
		zapcore.AddSync(&lumberjack.Logger{
			Filename:   cfg.AppFile,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
		}),
		zap.DebugLevel,
	)

	return &Log{
		access: zap.New(accessCore, zap.AddCaller()),
		err:    zap.New(errorCore, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel)),
		app:    zap.New(appCore, zap.AddCaller()),
	}
}

func (l *Log) Access(msg string, fields ...zap.Field) {
	l.access.Info(msg, fields...)
}

func (l *Log) Error(err error, msg string, fields ...zap.Field) {
	l.err.Error(msg, append(fields, zap.Error(err))...)
}

func (l *Log) App(msg string, fields ...zap.Field) {
	l.app.Info(msg, fields...)
}
