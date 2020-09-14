package common

import (
	"log"
	"os"
)

const (
	debug = iota
	info
	error
	fatal
)

var Log *logger

type logger struct {
	log   *log.Logger
	level int
}

func initLog() {
	// log.SetFlags(log.LstdFlags | log.Lshortfile)
	l := log.New(os.Stderr, "", log.LstdFlags|log.Lshortfile)
	Log = &logger{log: l, level: logLevel2Int(ServerCfg.LogLevel)}
}

func logLevel2Int(l string) int {
	switch l {
	case "debug":
		return debug
	case "info":
		return info
	case "error":
		return error
	case "fatal":
		return fatal
	default:
		return info
	}
}

func (l *logger) Debug(v ...interface{}) {
	if l.level > debug {
		return
	}
	data := append([]interface{}{"[Debug]"}, v...)
	l.log.Println(data...)
}

func (l *logger) Info(v ...interface{}) {
	if l.level > info {
		return
	}
	data := append([]interface{}{"[Info]"}, v...)
	l.log.Println(data...)
}

func (l *logger) Error(v ...interface{}) {
	if l.level > error {
		return
	}
	data := append([]interface{}{"[Error]"}, v...)
	l.log.Println(data...)
}

func (l *logger) Fatal(v ...interface{}) {
	if l.level > fatal {
		return
	}
	data := append([]interface{}{"[Fatal]"}, v...)
	l.log.Fatalln(data...)
}
