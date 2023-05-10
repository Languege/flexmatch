// dateformat
// @author LanguageY++2013 2023/5/10 11:00
// @company soulgame
package logger

import (
	"fmt"
	"github.com/Languege/flexmatch/common/logger/dateformat"
	"github.com/juju/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net/url"
	"strconv"
	"time"
)

func dateformatSinkBuildOptions(cfg *dateformat.Configure, queryArgs url.Values) (err error) {
	if v := queryArgs.Get("maxage"); v != "" {
		cfg.MaxAge, err = strconv.Atoi(v)
		if err != nil {
			err = errors.Trace(fmt.Errorf("maxage %s atoi err %v", v, err))
			return
		}
	}

	if v := queryArgs.Get("prefix"); v != "" {
		cfg.Prefix = v
	}

	return
}

func RegisterDateformatSink() {
	err := zap.RegisterSink("dateformat", func(u *url.URL) (zap.Sink, error) {
		cfg := dateformat.Configure{
			Dir: u.Host + u.Path,
		}

		if err := dateformatSinkBuildOptions(&cfg, u.Query()); err != nil {
			return nil, err
		}

		useCache, buffOpts, err := buildBufferedOptions(u.Query())
		if err != nil {
			return nil, err
		}

		if useCache {
			return &bufferedDateFormatSink{
				BufferedWriteSyncer: *dateformat.NewBufferedWriteSyncer(cfg, buffOpts...),
			}, nil
		}

		return dateformat.NewLogger(cfg), nil
	})

	FatalfIf(err != nil, "dateformat sink register err %v", err)
}

func buildBufferedOptions(queryArgs url.Values) (useCache bool, opts []dateformat.BufferedOption, err error) {
	if v := queryArgs.Get("usecache"); v != "" {
		useCache, err = strconv.ParseBool(v)
		if err != nil {
			return
		}
	}

	if !useCache {
		return
	}

	if v := queryArgs.Get("cachesize"); v != "" {
		//设置了写缓存，单位KB
		var size int
		size, err = strconv.Atoi(v)
		if err != nil {
			return
		}

		opts = append(opts, dateformat.WithBufferSize(size))
	}

	if v := queryArgs.Get("flushinterval"); v != "" {
		//设置了写缓存，单位KB
		var flushInterval time.Duration
		flushInterval, err = time.ParseDuration(v)
		if err != nil {
			return
		}

		opts = append(opts, dateformat.WithFlushInterval(flushInterval))
	}

	return
}

type bufferedDateFormatSink struct {
	zapcore.BufferedWriteSyncer
}

func (b *bufferedDateFormatSink) Close() error {
	return b.Stop()
}
