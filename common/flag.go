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
	passwd     string
	// 显示版本信息
	rev bool
)

func initFlag() {
	flag.StringVar(&serverFile, "conf", "./conf/server.toml", "server config file path")
	flag.StringVar(&passwd, "pass", "", "generation a sha1 password")
	flag.BoolVar(&rev, "rev", false, "display version info")
	flag.Parse()

	if passwd != "" {
		pwdHash := hashPass(passwd)
		fmt.Printf("passwd-sha1:%s\n", pwdHash)
		os.Exit(0)
	}

	if rev {
		fmt.Printf("%s v%s build on %s [%s, %s] commit_id(%s) \n",
			APP_NAME, APP_VER, runtime.Version(), runtime.GOOS, runtime.GOARCH, CommitId)
		os.Exit(0)
	}
}

func InitConfig() {
	initFlag()
	loadServer()
	loadUser()
	initIpPool()
}
