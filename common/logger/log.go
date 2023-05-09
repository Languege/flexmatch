// logger
// @author LanguageY++2013 2023/5/6 14:34
// @company soulgame
package logger

import (
	_ "github.com/Languege/flexmatch/service/match/conf"
	"go.uber.org/zap"
	"log"
	"go.uber.org/zap/zapcore"
)


var (
	logger Logger
)

type LoggerConfig struct {
	Level       string   `mapstructure:"level"`
	Paths       []string `mapstructure:"paths"`
	Encoding    string   `mapstructure:"encoding"`
	Development bool     `mapstructure:"development"`
}

func NewZapLogger(cfg LoggerConfig) (*zap.Logger, error) {
	config := zap.NewProductionConfig()
	err := config.Level.UnmarshalText([]byte(cfg.Level))
	if err != nil {
		return nil, err
	}

	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	config.OutputPaths = cfg.Paths
	//development模式下DPanic直接panic
	config.Development = cfg.Development


	p, err := config.Build(zap.AddCallerSkip(1))
	if err != nil {
		return nil, err
	}

	return p, nil
}

//日志初始化
func LoadConfig(cfg LoggerConfig) {
	RegisterSinkLumberjackSink()

	p, err := NewZapLogger(cfg)
	FatalIfError(err)

	logger = p.Sugar()
}

// Logger logger interface
type Logger interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Debugw(msg string, keysAndValues ...interface{})

	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Infow(msg string, keysAndValues ...interface{})

	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Warnw(msg string, keysAndValues ...interface{})

	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Errorw(msg string, keysAndValues ...interface{})

	DPanic(args ...interface{})
	DPanicf(format string, args ...interface{})
	DPanicw(msg string, keysAndValues ...interface{})

	Panic(v ...interface{})
	Panicf(format string, args ...interface{})
	Panicw(msg string, keysAndValues ...interface{})

	Fatal(v ...interface{})
	Fatalf(format string, args ...interface{})
	Fatalw(msg string, keysAndValues ...interface{})
}

// Debug log to level debug
func Debug(args ...interface{}) {
	logger.Debug(args...)
}

// Debugf log to level debug
func Debugf(fmt string, args ...interface{}) {
	logger.Debugf(fmt, args...)
}

// Debugw log to level debug
func Debugw(msg string, keysAndValues ...interface{}) {
	logger.Debugw(msg, keysAndValues...)
}

// Info log to level info
func Info(args ...interface{}) {
	logger.Info(args...)
}

// Infof log to level info
func Infof(fmt string, args ...interface{}) {
	logger.Infof(fmt, args...)
}

// Infow log to level info
func Infow(msg string, keysAndValues ...interface{}) {
	logger.Infow(msg, keysAndValues...)
}

// Warn log to level warn
func Warn(args ...interface{}) {
	logger.Warn(args...)
}

// Warnf log to level warn
func Warnf(fmt string, args ...interface{}) {
	logger.Warnf(fmt, args...)
}

// Warnw log to level warn
func Warnw(msg string, keysAndValues ...interface{}) {
	logger.Warnf(msg, keysAndValues...)
}


// Error log to level error
func Error(args ...interface{}) {
	logger.Error(args...)
}

// Errorf log to level error
func Errorf(fmt string, args ...interface{}) {
	logger.Errorf(fmt, args...)
}

// Errorw log to level error
func Errorw(msg string, keysAndValues ...interface{}) {
	logger.Errorw(msg, keysAndValues...)
}

// DPanic uses fmt.Sprint to construct and log a message. In development, the
// logger then panics. (See DPanicLevel for details.)
func DPanic(args ...interface{}) {
	logger.DPanic(args...)
}

// DPanicf uses fmt.Sprintf to log a templated message. In development, the
// logger then panics. (See DPanicLevel for details.)
func DPanicf(fmt string, args ...interface{}) {
	logger.DPanicf(fmt, args...)
}

// DPanicw logs a message with some additional context. In development, the
// logger then panics. (See DPanicLevel for details.) The variadic key-value
// pairs are treated as they are in With.
func DPanicw(msg string, keysAndValues ...interface{}) {
	logger.DPanicw(msg, keysAndValues...)
}

// Panic uses fmt.Sprint to construct and log a message, then panics.
func Panic(args ...interface{}) {
	logger.Panic(args...)
}

// Panicf uses fmt.Sprintf to log a templated message, then panics.
func Panicf(fmt string, args ...interface{}) {
	logger.Panicf(fmt, args...)
}

// Panicw logs a message with some additional context, then panics. The
// variadic key-value pairs are treated as they are in With.
func Panicw(msg string, keysAndValues ...interface{}) {
	logger.Panicw(msg, keysAndValues...)
}

// Fatal uses fmt.Sprint to construct and log a message, then panics.
func Fatal(args ...interface{}) {
	logger.Fatal(args...)
}

// Fatalf uses fmt.Sprintf to log a templated message, then panics.
func Fatalf(format string, args ...interface{}) {
	logger.Fatalf(format, args...)
}

// Fatalw logs a message with some additional context, then panics. The
// variadic key-value pairs are treated as they are in With.
func Fatalw(msg string, keysAndValues ...interface{}) {
	logger.Fatalw(msg, keysAndValues...)
}


// FatalfIf log to level error
func FatalfIf(cond bool, fmt string, args ...interface{}) {
	if !cond {
		return
	}
	log.Fatalf(fmt, args...)
}

// FatalIfError if err is not nil, then log to level fatal and call os.Exit
func FatalIfError(err error) {
	FatalfIf(err != nil, "fatal error: %v", err)
}