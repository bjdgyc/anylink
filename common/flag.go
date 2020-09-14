package common

import (
	"flag"
	"fmt"
	"os"
	"runtime"
)

var (
	// 提交id
	CommitId string
	// 配置文件
	serverFile string
	// 显示版本信息
	rev bool
)

func initFlag() {
	flag.StringVar(&serverFile, "conf", "./conf/server.toml", "server config file path")
	flag.BoolVar(&rev, "rev", false, "display version info")
	flag.Parse()

	if rev {
		fmt.Printf("%s v%s build on %s [%s, %s] commit_id(%s) \n",
			APP_NAME, APP_VER, runtime.Version(), runtime.GOOS, runtime.GOARCH, CommitId)
		os.Exit(0)
	}
}

func InitConfig() {
	initFlag()
	loadServer()
	initLog()
}
