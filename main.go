package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bjdgyc/anylink/common"
	"github.com/bjdgyc/anylink/handler"
)

var COMMIT_ID string

func main() {
	log.Println("start")
	common.CommitId = COMMIT_ID
	common.InitConfig()
	handler.Start()
	signalWatch()
}

func signalWatch() {
	fmt.Println("Server pid: ", os.Getpid())

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGALRM, syscall.SIGUSR2)
	for {
		sig := <-sigs
		fmt.Printf("Get signal: %v \n", sig)
		switch sig {
		case syscall.SIGUSR2:
			// reload
			fmt.Println("reload")
		default:
			// stop
			handler.Stop()
			return
		}
	}
}
