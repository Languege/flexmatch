// logger
// @author LanguageY++2013 2023/5/6 14:34
// @company soulgame
package logger

import (
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"net/url"
	"encoding/json"
	"log"
	"go.uber.org/zap/zapcore"
	"strings"
	"fmt"
	"github.com/spf13/viper"
	"path/filepath"
)



const (
	// StdErr is the default configuration for log output.
	StdErr = "stderr"
	// StdOut configuration for log output
	StdOut = "stdout"
)

var (
	logger Logger
)

func init() {
	InitLog("debug", false)
}

// Logger logger interface
type Logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

// InitLog is an initialization for a logger
// level can be: debug info warn error
func InitLog(level string, debug bool) {
	logDir, err := filepath.Abs(viper.GetString("log.path"))
	FatalIfError(err)

	InitLog2(level, logDir + "/runtime.log", 1, `{"filename":"runtime.log","MaxAge":7}`, debug)
}

// InitLog2 specify advanced log config
func InitLog2(level string, outputs string, logRotationEnable int64, logRotateConfigJSON string, debug bool) {
	outputPaths := strings.Split(outputs, ",")
	for i, v := range outputPaths {
		if logRotationEnable != 0 && v != StdErr && v != StdOut {
			outputPaths[i] = fmt.Sprintf("lumberjack://%s", v)
		}
	}

	if logRotationEnable != 0 {
		setupLogRotation(logRotateConfigJSON)
	}

	config := loadConfig(level, debug)
	config.OutputPaths = outputPaths
	p, err := config.Build(zap.AddCallerSkip(1))
	FatalIfError(err)
	logger = p.Sugar()
}


type lumberjackSink struct {
	lumberjack.Logger
}

func (*lumberjackSink) Sync() error {
	return nil
}

// setupLogRotation initializes log rotation for a single file path target.
func setupLogRotation(logRotateConfigJSON string) {
	err := zap.RegisterSink("lumberjack", func(u *url.URL) (zap.Sink, error) {
		var conf lumberjackSink
		err := json.Unmarshal([]byte(logRotateConfigJSON), &conf)
		FatalfIf(err != nil, "bad config LogRotateConfigJSON: %v", err)
		conf.Filename = u.Host + u.Path
		return &conf, nil
	})
	FatalIfError(err)
}

func loadConfig(logLevel string, debug bool) zap.Config {
	config := zap.NewProductionConfig()
	err := config.Level.UnmarshalText([]byte(logLevel))
	FatalIfError(err)
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	if debug {
		config.Encoding = "console"
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	return config
}

// Debugf log to level debug
func Debugf(fmt string, args ...interface{}) {
	logger.Debugf(fmt, args...)
}

// Infof log to level info
func Infof(fmt string, args ...interface{}) {
	logger.Infof(fmt, args...)
}

// Warnf log to level warn
func Warnf(fmt string, args ...interface{}) {
	logger.Warnf(fmt, args...)
}

// Errorf log to level error
func Errorf(fmt string, args ...interface{}) {
	logger.Errorf(fmt, args...)
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