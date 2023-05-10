// logger
// @author LanguageY++2013 2023/5/8 16:41
// @company soulgame
package logger

import (
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"net/url"
	"strconv"
	"github.com/juju/errors"
	"fmt"
)

type lumberjackSink struct {
	lumberjack.Logger
}

func (*lumberjackSink) Sync() error {
	return nil
}

func(s *lumberjackSink) BuildOptions(queryArgs url.Values) (err error) {
	if v := queryArgs.Get("maxage"); v != "" {
		s.MaxAge, err = strconv.Atoi(v)
		if err != nil {
			err = errors.Trace(fmt.Errorf("maxage %s atoi err %v", v, err))
			return
		}
	}

	if v := queryArgs.Get("maxsize"); v != "" {
		s.MaxSize, err = strconv.Atoi(v)
		if err != nil {
			err = errors.Trace(fmt.Errorf("maxsize %s atoi err %v", v, err))
			return
		}
	}

	if v := queryArgs.Get("maxbackups"); v != "" {
		s.MaxBackups, err = strconv.Atoi(v)
		if err != nil {
			err = errors.Trace(fmt.Errorf("maxbackups %s atoi err %v", v, err))
			return
		}
	}

	return
}

func RegisterSinkLumberjackSink() {
	err := zap.RegisterSink("lumberjack", func(u *url.URL) (zap.Sink, error) {
		s := &lumberjackSink{}
		s.Filename = u.Host + u.Path
		return s, s.BuildOptions(u.Query())
	})

	FatalfIf(err != nil, "lumberjack sink register err %v", err)
}