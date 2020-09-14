package common

import "fmt"

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

func HumanByte(bAll float64) string {
	var hb string

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
