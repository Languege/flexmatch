// dateformat
// @author LanguageY++2013 2023/5/10 14:44
// @company soulgame
package dateformat

import (
	"testing"
	"strings"
	"time"
)

func TestLogger_Write(t *testing.T) {
	logger := NewLogger(Configure{
		Dir: "./logs",
	})

	//10KB大小
	logger.Write([]byte(strings.Repeat("s", 1024 * 10) + "\n"))
	//1KB大小
	logger.Write([]byte(strings.Repeat("s", 1024) + "\n"))
}

func TestNewBufferedWriteSyncer(t *testing.T) {
	flushTime := time.Second * 10
	logger := NewBufferedWriteSyncer(Configure{
		Dir: "./logs",
	}, WithBufferSize(10 * 1024), WithFlushInterval(flushTime))

	//defer logger.Stop()

	//1m大小
	logger.Write([]byte(strings.Repeat("s", 1024 * 10) + "\n"))
	//1kb大小
	logger.Write([]byte(strings.Repeat("s", 1024) + "\n"))

	time.Sleep(flushTime)
}
