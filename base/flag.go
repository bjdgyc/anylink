package base

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/bjdgyc/anylink/pkg/utils"
)

var (
	// 提交id
	CommitId string
	// 配置文件
	serverFile string
	// pass明文
	passwd string
	// 显示版本信息
	rev bool
)

func initFlag() {
	flag.StringVar(&serverFile, "conf", "./conf/server.toml", "server config file path")
	flag.StringVar(&passwd, "passwd", "", "the password plaintext")
	flag.BoolVar(&rev, "rev", false, "display version info")
	flag.Parse()

	if passwd != "" {
		pass, _ := utils.PasswordHash(passwd)
		fmt.Printf("Passwd:%s\n", pass)
		os.Exit(0)
	}

	if rev {
		fmt.Printf("%s v%s build on %s [%s, %s] commit_id(%s) \n",
			APP_NAME, APP_VER, runtime.Version(), runtime.GOOS, runtime.GOARCH, CommitId)
		os.Exit(0)
	}
}
