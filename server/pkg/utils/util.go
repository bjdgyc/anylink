package utils

import (
	"fmt"
	"math/rand"
	"strings"
	"sync/atomic"
	"time"
)

var (
	// 每秒时间缓存
	timeNowSec = &atomic.Value{}
)

func init() {
	rand.Seed(time.Now().UnixNano())

	timeNowSec.Store(time.Now())
	go func() {
		tick := time.NewTicker(time.Second * 1)
		for c := range tick.C {
			timeNowSec.Store(c)
		}
	}()
}

func NowSec() time.Time {
	t := timeNowSec.Load()
	return t.(time.Time)
}

func InArrStr(arr []string, str string) bool {
	for _, d := range arr {
		if d == str {
			return true
		}
	}
	return false
}

const (
	KB = 1024
	MB = 1024 * KB
	GB = 1024 * MB
	TB = 1024 * GB
	PB = 1024 * TB
)

func HumanByte(bf interface{}) string {
	var hb string
	var bAll float64
	switch bi := bf.(type) {
	case int:
		bAll = float64(bi)
	case int32:
		bAll = float64(bi)
	case uint32:
		bAll = float64(bi)
	case int64:
		bAll = float64(bi)
	case uint64:
		bAll = float64(bi)
	case float64:
		bAll = float64(bi)
	}

	switch {
	case bAll >= TB:
		hb = fmt.Sprintf("%0.2f TB", bAll/TB)
	case bAll >= GB:
		hb = fmt.Sprintf("%0.2f GB", bAll/GB)
	case bAll >= MB:
		hb = fmt.Sprintf("%0.2f MB", bAll/MB)
	case bAll >= KB:
		hb = fmt.Sprintf("%0.2f KB", bAll/KB)
	default:
		hb = fmt.Sprintf("%0.2f B", bAll)
	}

	return hb
}

func RandomRunes(length int) string {
	letterRunes := []rune("abcdefghijklmnpqrstuvwxy1234567890")

	bytes := make([]rune, length)

	for i := range bytes {
		bytes[i] = letterRunes[rand.Intn(len(letterRunes))]
	}

	return string(bytes)
}

func ParseName(name string) string {
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, "'", "-")
	name = strings.ReplaceAll(name, "\"", "-")
	name = strings.ReplaceAll(name, ";", "-")
	return name
}
