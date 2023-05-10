// dateformat
// @author LanguageY++2013 2023/5/10 16:33
// @company soulgame
package dateformat

import (
	"go.uber.org/zap/zapcore"
	"time"
)

type BufferedOption func(b *zapcore.BufferedWriteSyncer)

func WithBufferSize(size int) BufferedOption {
	return func(b *zapcore.BufferedWriteSyncer) {
		b.Size = size
	}
}

func WithFlushInterval(flushInterval time.Duration) BufferedOption {
	return func(b *zapcore.BufferedWriteSyncer) {
		b.FlushInterval = flushInterval
	}
}

func NewBufferedWriteSyncer(cfg Configure, opts... BufferedOption) *zapcore.BufferedWriteSyncer {
	buffWS := &zapcore.BufferedWriteSyncer{
		WS: NewLogger(cfg),
	}

	for _, opt := range opts {
		opt(buffWS)
	}

	return buffWS
}
