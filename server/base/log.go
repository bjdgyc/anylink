package base

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"time"
)

const (
	LogLevelTrace = iota
	LogLevelDebug
	LogLevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelFatal
)

var (
	baseLw    *logWriter
	baseLog   *log.Logger
	baseLevel int
	levels    map[int]string

	dateFormat = "2006-01-02"
	logName    = "anylink.log"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

// 实现 os.Writer 接口
type logWriter struct {
	UseStdout bool
	FileName  string
	File      *os.File
	NowDate   string
}

// 实现日志文件的切割
func (lw *logWriter) Write(p []byte) (n int, err error) {
	if !lw.UseStdout {
		return lw.File.Write(p)
	}

	date := time.Now().Format(dateFormat)
	if lw.NowDate != date {
		_ = lw.File.Close()
		_ = os.Rename(lw.FileName, lw.FileName+"."+lw.NowDate)
		lw.NowDate = date
		lw.newFile()
	}
	return lw.File.Write(p)
}

// 创建新文件
func (lw *logWriter) newFile() {
	if lw.UseStdout {
		lw.File = os.Stdout
		return
	}

	f, err := os.OpenFile(lw.FileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	lw.File = f
}

func initLog() {
	// 初始化 baseLog
	baseLw = &logWriter{
		UseStdout: Cfg.LogPath == "",
		FileName:  path.Join(Cfg.LogPath, logName),
		NowDate:   time.Now().Format(dateFormat),
	}

	baseLw.newFile()
	baseLevel = logLevel2Int(Cfg.LogLevel)
	baseLog = log.New(baseLw, "", log.LstdFlags|log.Lshortfile)

	serverLog = log.New(&sLogWriter{}, "[http_server]", log.LstdFlags|log.Lshortfile)
}

func GetBaseLw() *logWriter {
	return baseLw
}

var serverLog *log.Logger

type sLogWriter struct{}

func (w *sLogWriter) Write(p []byte) (n int, err error) {
	if Cfg.HttpServerLog {
		return os.Stderr.Write(p)
	}
	return 0, nil
}

// 获取 log.Logger
func GetServerLog() *log.Logger {
	return serverLog
}

func GetLogLevel() int {
	return baseLevel
}

func logLevel2Int(l string) int {
	levels = map[int]string{
		LogLevelTrace: "Trace",
		LogLevelDebug: "Debug",
		LogLevelInfo:  "Info",
		LogLevelWarn:  "Warn",
		LogLevelError: "Error",
		LogLevelFatal: "Fatal",
	}
	lvl := LogLevelInfo
	for k, v := range levels {
		if strings.ToLower(l) == strings.ToLower(v) {
			lvl = k
		}
	}
	return lvl
}

func output(l int, s ...interface{}) {
	lvl := fmt.Sprintf("[%s] ", levels[l])
	_ = baseLog.Output(3, lvl+fmt.Sprintln(s...))
}

func Trace(v ...interface{}) {
	l := LogLevelTrace
	if baseLevel > l {
		return
	}
	output(l, v...)
}

func Debug(v ...interface{}) {
	l := LogLevelDebug
	if baseLevel > l {
		return
	}
	output(l, v...)
}

func Info(v ...interface{}) {
	l := LogLevelInfo
	if baseLevel > l {
		return
	}
	output(l, v...)
}

func Warn(v ...interface{}) {
	l := LogLevelWarn
	if baseLevel > l {
		return
	}
	output(l, v...)
}

func Error(v ...interface{}) {
	l := LogLevelError
	if baseLevel > l {
		return
	}
	output(l, v...)
}

func Fatal(v ...interface{}) {
	l := LogLevelFatal
	if baseLevel > l {
		return
	}
	output(l, v...)
	os.Exit(1)
}
