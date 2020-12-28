package base

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/bjdgyc/anylink/pkg/utils"
)

var (
	// 提交id
	CommitId string
	// 配置文件
	serverFile string
	// pass明文
	passwd string
	// 生成密钥
	secret bool
	// 显示版本信息
	rev bool
)

func initFlag() {
	flag.StringVar(&serverFile, "conf", "./conf/server.toml", "server config file path")
	flag.StringVar(&passwd, "passwd", "", "convert the password plaintext")
	flag.BoolVar(&secret, "secret", false, "generate a random jwt secret")
	flag.BoolVar(&rev, "rev", false, "display version info")
	flag.Parse()

	if passwd != "" {
		pass, _ := utils.PasswordHash(passwd)
		fmt.Printf("Passwd:%s\n", pass)
		os.Exit(0)
	}

	if secret {
		rand.Seed(time.Now().UnixNano())
		s, _ := utils.RandSecret(40, 60)
		s = strings.Trim(s, "=")
		fmt.Printf("Secret:%s\n", s)
		os.Exit(0)
	}

	if rev {
		fmt.Printf("%s v%s build on %s [%s, %s] commit_id(%s) \n",
			APP_NAME, APP_VER, runtime.Version(), runtime.GOOS, runtime.GOARCH, CommitId)
		os.Exit(0)
	}
}
