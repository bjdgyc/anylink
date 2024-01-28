// AnyLink 是一个企业级远程办公vpn软件，可以支持多人同时在线使用。

//go:build !windows
// +build !windows

package main

import (
	"embed"
	"os"
	"os/signal"
	"syscall"

	"github.com/bjdgyc/anylink/admin"
	"github.com/bjdgyc/anylink/base"
	"github.com/bjdgyc/anylink/handler"
)

//go:embed ui
var uiData embed.FS

// 程序版本
var (
	appVer   string
	commitId string
	date     string
)

func main() {
	admin.UiData = uiData
	base.APP_VER = appVer
	base.CommitId = commitId
	base.Date = date

	base.Start()
	handler.Start()

	signalWatch()
}

func signalWatch() {
	base.Info("Server pid: ", os.Getpid())

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGALRM, syscall.SIGUSR2)
	for {
		sig := <-sigs
		base.Info("Get signal:", sig)
		switch sig {
		case syscall.SIGUSR2:
			// reload
			base.Info("Reload")
		default:
			// stop
			base.Info("Stop")
			handler.Stop()
			return
		}
	}
}
