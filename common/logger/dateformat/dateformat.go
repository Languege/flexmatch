// dateformat
// @author LanguageY++2013 2023/5/10 09:36
// @company soulgame
package dateformat

import (
	"fmt"
	"os"
	"sync"
	"time"
	"strings"
	"errors"
	"sort"
	"path/filepath"
	"log"
)

const (
	timeFormat = "2006010215"
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

	millCh    chan bool
	startMill sync.Once
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
	l.mill()

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
	dateformat := time.Now().Format(timeFormat)
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

//发送旧日志的移除的信号，首次触发时创建信息接收处理协程
func (l *Logger) mill() {
	l.startMill.Do(func() {
		l.millCh = make(chan bool, 1)
		go l.millRun()
	})
	select {
	case l.millCh <- true:
	default:
	}
}

func (l *Logger) millRun() {
	for  range l.millCh {
		err := l.millRunOnce()
		if err != nil {
			log.Printf("mill run once err %s\n", err)
		}
	}
}

//旧日志移除处理
func (l *Logger) millRunOnce() error {
	files, err := l.oldLogFiles()
	if err != nil {
		return err
	}
	var remove []logInfo
	if l.MaxAge > 0 {
		diff := time.Duration(int64(24*time.Hour) * int64(l.MaxAge))
		cutoff := time.Now().Add(-1 * diff)

		for _, f := range files {
			if f.timestamp.Before(cutoff) {
				remove = append(remove, f)
			}
		}
	}

	for _, f := range remove {
		errRemove := os.Remove(filepath.Join(l.Dir, f.Name()))
		if err == nil && errRemove != nil {
			err = errRemove
		}
	}

	return err
}

func (l *Logger) oldLogFiles() ([]logInfo, error) {
	files, err := os.ReadDir(l.Dir)
	if err != nil {
		return nil, fmt.Errorf("can't read log file directory: %s", err)
	}
	logFiles := []logInfo{}

	prefix := ""
	if l.Prefix != "" {
		prefix = l.Prefix + "-"
	}
	ext := ".log"

	for _, f := range files {
		if f.IsDir() {
			continue
		}
		if t, err := l.timeFromName(f.Name(), prefix, ext); err == nil {
			logFiles = append(logFiles, logInfo{t, f})
			continue
		}
	}

	sort.Sort(byFormatTime(logFiles))

	return logFiles, nil
}

func (l *Logger) timeFromName(filename, prefix, ext string) (time.Time, error) {
	if  prefix != "" && !strings.HasPrefix(filename, prefix) {
		return time.Time{}, errors.New("mismatched prefix")
	}
	if !strings.HasSuffix(filename, ext) {
		return time.Time{}, errors.New("mismatched extension")
	}
	ts := filename[len(prefix) : len(filename)-len(ext)]
	return time.Parse(timeFormat, ts)
}





// logInfo is a convenience struct to return the filename and its embedded
// timestamp.
type logInfo struct {
	timestamp time.Time
	os.DirEntry
}

// byFormatTime sorts by newest time formatted in the name.
type byFormatTime []logInfo

func (b byFormatTime) Less(i, j int) bool {
	return b[i].timestamp.After(b[j].timestamp)
}

func (b byFormatTime) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b byFormatTime) Len() int {
	return len(b)
}
