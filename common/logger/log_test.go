// logger
// @author LanguageY++2013 2023/5/8 16:08
// @company soulgame
package logger

import (
	"testing"
	"go.uber.org/zap"
	"net/url"
	"sync/atomic"
	"strings"
	"time"
)

type testSink struct {
	zap.Sink
	wCount int64
	syncCount int64

	filename string
}

func(s *testSink) Write(p []byte) (n int, err error) {
	atomic.AddInt64(&s.wCount, 1)
	return len(p), nil
}

func(s *testSink) Sync() error {
	atomic.AddInt64(&s.syncCount, 1)
	return nil
}

func(s *testSink) Close() error {
	return nil
}

func TestDebug(t *testing.T) {
	s := &testSink{}
	zap.RegisterSink("test", func(u *url.URL) (zap.Sink, error) {
		s.filename = u.Path
		params := u.Query()
		t.Log(params)
		return s, nil
	})

	cfg := zap.NewProductionConfig()

	cfg.OutputPaths = []string{"test:///./test.log?maxage=1d"}
	log, err := cfg.Build(zap.AddCallerSkip(1))
	if err != nil {
		t.Fatal(err)
	}

	slog := log.Sugar()

	N := 1
	t.Run("count", func(t *testing.T) {
		for i := 0; i < N; i++ {
			slog.Infof("info %d", i)
			//should not log
			slog.Debugf("debug %d", i)
		}
	})

	if s.wCount != int64(N) {
		t.Fatalf("wCount should equal %d", N)
	}
}

func init() {
	RegisterDateformatSink()
}
func TestDateformatSink_BuildOptions(t *testing.T) {
	cfg := zap.NewProductionConfig()

	cfg.OutputPaths = []string{"dateformat://./test.log?usecache=true&cachesize=10"}

	log, err := cfg.Build(zap.AddCallerSkip(1))
	if err != nil {
		t.Fatal(err)
	}
	defer log.Sync()

	slog := log.Sugar()

	slog.Warn(strings.Repeat("a", 1024 * 10))

	slog.Warn(strings.Repeat("a", 1024))
}

func TestDateformatSink_Filename(t *testing.T) {
	cfg := zap.NewProductionConfig()

	cfg.OutputPaths = []string{"dateformat://./test.log", "stdout"}

	log, err := cfg.Build(zap.AddCallerSkip(1))
	if err != nil {
		t.Fatal(err)
	}
	defer log.Sync()

	slog := log.Sugar()

	tick := time.Tick(time.Second)
	for now := range tick {
		slog.Info(now.String())
	}
}


//-test.benchtime=20s
//BenchmarkRegisterDateformatSink-8   	28299568	       801.8 ns/op
func BenchmarkRegisterDateformatSink(b *testing.B) {
	cfg := zap.NewProductionConfig()

	cfg.OutputPaths = []string{"dateformat://./test.log"}

	log, err := cfg.Build(zap.AddCallerSkip(1))
	if err != nil {
		b.Fatal(err)
	}
	defer log.Sync()

	data := strings.Repeat("a", 512)
	for i := 0; i < b.N; i++ {
		log.Info(data)
	}
}

//-test.benchtime=20s
//BenchmarkRegisterDateformatSinkWithBuffer-8   	29431770	       813.7 ns/op
func BenchmarkRegisterDateformatSinkWithBuffer(b *testing.B) {
	cfg := zap.NewProductionConfig()

	cfg.OutputPaths = []string{"dateformat://./test.log?usecache=true&cachesize=256&flushinterval=5s"}

	log, err := cfg.Build(zap.AddCallerSkip(1))
	if err != nil {
		b.Fatal(err)
	}
	defer log.Sync()

	data := strings.Repeat("a", 512)
	for i := 0; i < b.N; i++ {
		log.Info(data)
	}
}