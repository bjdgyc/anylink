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
	_Debug = iota
	_Info
	_Warn
	_Error
	_Fatal
)

var (
	baseLog   *log.Logger
	baseLevel int
	levels    map[int]string

	dateFormat = "2006-01-02"
	logName    = "anylink.log"
)

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
	baseLw := &logWriter{
		UseStdout: Cfg.LogPath == "",
		FileName:  path.Join(Cfg.LogPath, logName),
		NowDate:   time.Now().Format(dateFormat),
	}

	baseLw.newFile()
	baseLevel = logLevel2Int(Cfg.LogLevel)
	baseLog = log.New(baseLw, "", log.LstdFlags|log.Lshortfile)
}

// 获取 log.Logger
func GetBaseLog() *log.Logger {
	return baseLog
}

func logLevel2Int(l string) int {
	levels = map[int]string{
		_Debug: "Debug",
		_Info:  "Info",
		_Warn:  "Warn",
		_Error: "Error",
		_Fatal: "Fatal",
	}
	lvl := _Info
	for k, v := range levels {
		if strings.EqualFold(strings.ToLower(l), strings.ToLower(v)) {
			lvl = k
		}
	}
	return lvl
}

func output(l int, s ...interface{}) {
	lvl := fmt.Sprintf("[%s] ", levels[l])
	_ = baseLog.Output(3, lvl+fmt.Sprintln(s...))
}

func Debug(v ...interface{}) {
	l := _Debug
	if baseLevel > l {
		return
	}
	output(l, v...)
}

func Info(v ...interface{}) {
	l := _Info
	if baseLevel > l {
		return
	}
	output(l, v...)
}

func Warn(v ...interface{}) {
	l := _Warn
	if baseLevel > l {
		return
	}
	output(l, v...)
}

func Error(v ...interface{}) {
	l := _Error
	if baseLevel > l {
		return
	}
	output(l, v...)
}

func Fatal(v ...interface{}) {
	l := _Fatal
	if baseLevel > l {
		return
	}
	output(l, v...)
	os.Exit(1)
}
