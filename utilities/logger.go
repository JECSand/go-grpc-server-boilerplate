package utilities

import (
	"github.com/JECSand/go-grpc-server-boilerplate/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

// Logger methods interface
type Logger interface {
	InitLogger()
	Debug(args ...interface{})
	Debugf(template string, args ...interface{})
	Info(args ...interface{})
	Infof(template string, args ...interface{})
	Warn(args ...interface{})
	Warnf(template string, args ...interface{})
	Error(args ...interface{})
	Errorf(template string, args ...interface{})
	DPanic(args ...interface{})
	DPanicf(template string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(template string, args ...interface{})
	Printf(template string, args ...interface{})
}

// Logger
type apiLogger struct {
	cfg         *config.Configuration
	sugarLogger *zap.SugaredLogger
}

// NewAPILogger Logger constructor
func NewAPILogger(cfg *config.Configuration) *apiLogger {
	return &apiLogger{cfg: cfg}
}

// For mapping config logger to email_service logger levels
var loggerLevelMap = map[string]zapcore.Level{
	"debug":  zapcore.DebugLevel,
	"info":   zapcore.InfoLevel,
	"warn":   zapcore.WarnLevel,
	"error":  zapcore.ErrorLevel,
	"dpanic": zapcore.DPanicLevel,
	"panic":  zapcore.PanicLevel,
	"fatal":  zapcore.FatalLevel,
}

func (l *apiLogger) getLoggerLevel(cfg *config.Configuration) zapcore.Level {
	level, exist := loggerLevelMap[cfg.Logger.Level]
	if !exist {
		return zapcore.DebugLevel
	}
	return level
}

// InitLogger Init logger
func (l *apiLogger) InitLogger() {
	logLevel := l.getLoggerLevel(l.cfg)
	logWriter := zapcore.AddSync(os.Stderr)
	var encoderCfg zapcore.EncoderConfig
	if l.cfg.ENV == "production" {
		encoderCfg = zap.NewProductionEncoderConfig()
	} else {
		encoderCfg = zap.NewDevelopmentEncoderConfig()
	}
	var encoder zapcore.Encoder
	encoderCfg.LevelKey = "LEVEL"
	encoderCfg.CallerKey = "CALLER"
	encoderCfg.TimeKey = "TIME"
	encoderCfg.NameKey = "NAME"
	encoderCfg.MessageKey = "MESSAGE"
	if l.cfg.Logger.Encoding == "console" {
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	}
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	core := zapcore.NewCore(encoder, logWriter, zap.NewAtomicLevelAt(logLevel))
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	l.sugarLogger = logger.Sugar()
	if err := l.sugarLogger.Sync(); err != nil {
		l.sugarLogger.Error(err)
	}
}

// Debug Logger method
func (l *apiLogger) Debug(args ...interface{}) {
	l.sugarLogger.Debug(args...)
}

// Debugf Logger method
func (l *apiLogger) Debugf(template string, args ...interface{}) {
	l.sugarLogger.Debugf(template, args...)
}

// Info Logger method
func (l *apiLogger) Info(args ...interface{}) {
	l.sugarLogger.Info(args...)
}

// Infof Logger method
func (l *apiLogger) Infof(template string, args ...interface{}) {
	l.sugarLogger.Infof(template, args...)
}

// Printf Logger method
func (l *apiLogger) Printf(template string, args ...interface{}) {
	l.sugarLogger.Infof(template, args...)
}

// Warn Logger method
func (l *apiLogger) Warn(args ...interface{}) {
	l.sugarLogger.Warn(args...)
}

// Warnf Logger method
func (l *apiLogger) Warnf(template string, args ...interface{}) {
	l.sugarLogger.Warnf(template, args...)
}

// Error Logger method
func (l *apiLogger) Error(args ...interface{}) {
	l.sugarLogger.Error(args...)
}

// Errorf Logger method
func (l *apiLogger) Errorf(template string, args ...interface{}) {
	l.sugarLogger.Errorf(template, args...)
}

// DPanic Logger method
func (l *apiLogger) DPanic(args ...interface{}) {
	l.sugarLogger.DPanic(args...)
}

// DPanicf Logger method
func (l *apiLogger) DPanicf(template string, args ...interface{}) {
	l.sugarLogger.DPanicf(template, args...)
}

// Panic Logger method
func (l *apiLogger) Panic(args ...interface{}) {
	l.sugarLogger.Panic(args...)
}

// Panicf Logger method
func (l *apiLogger) Panicf(template string, args ...interface{}) {
	l.sugarLogger.Panicf(template, args...)
}

// Fatal Logger method
func (l *apiLogger) Fatal(args ...interface{}) {
	l.sugarLogger.Fatal(args...)
}

// Fatalf Logger method
func (l *apiLogger) Fatalf(template string, args ...interface{}) {
	l.sugarLogger.Fatalf(template, args...)
}
