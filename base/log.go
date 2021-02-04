package base

import (
	"fmt"
	"log"
	"os"
	"strings"
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
)

func initLog() {
	baseLog = log.New(os.Stderr, "", log.LstdFlags|log.Lshortfile)
	baseLevel = logLevel2Int(Cfg.LogLevel)
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
