// dateformat
// @author LanguageY++2013 2023/5/10 09:36
// @company soulgame
package dateformat

import (
	"fmt"
	"os"
	"sync"
	"time"
)

const (
	cacheUnit = 1024
)

type Configure struct {
	//Dir 日志文件目录
	Dir string `json:"dir"`

	//Prefix 文件名前缀
	Prefix string `json:"prefix"`

	//MaxAge 日志保留天数
	MaxAge int `json:"maxage"`
}

type Logger struct {
	Configure

	file *os.File
	mu   sync.Mutex

	cache 	[]byte
	free int
}

func NewLogger(cfg Configure) *Logger {
	l := &Logger{
		Configure:cfg,
	}

	return l
}

func (l *Logger) Write(p []byte) (n int, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	//文件为空，或者文件名发生变更（时间format划分）
	if l.file == nil ||  l.file.Name() != l.filename() {
		if err = l.openExistingOrNew(); err != nil {
			return
		}
	}

	n, err = l.file.Write(p)
	return
}


func (l *Logger) Sync() error {
	return nil
}

func(l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.close()
}

func(l *Logger) close() error {
	if l.file != nil {
		return l.file.Close()
	}

	l.file = nil

	return nil
}

func (l *Logger) openExistingOrNew() (err error) {
	filename := l.filename()
	_, err = os.Stat(filename)
	if os.IsNotExist(err) {
		return l.openNew()
	}

	if err != nil {
		return fmt.Errorf("error getting log file info: %s", err)
	}

	//打开已有文件
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		// if we fail to open the old log file for some reason, just ignore
		// it and open a new log file.
		return l.openNew()
	}
	l.file = file

	return
}

func (l *Logger) filename() string {
	dateformat := time.Now().Format("2006010215")
	if l.Prefix == "" {
		return fmt.Sprintf("%s/%s.log", l.Dir, dateformat)
	}

	return fmt.Sprintf("%s/%s-%s.log", l.Dir, l.Prefix, dateformat)
}

func (l *Logger) openNew() error {
	err := os.MkdirAll(l.Dir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("can't make directorie %s : %s", l.Dir, err)
	}

	filename := l.filename()
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("can't open new logfile: %s", err)
	}

	l.file = file
	return nil
}
