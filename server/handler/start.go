package handler

import (
	"crypto/sha1"
	"encoding/hex"
	"log"
	"os"

	"github.com/bjdgyc/anylink/admin"
	"github.com/bjdgyc/anylink/base"
	"github.com/bjdgyc/anylink/cron"
	"github.com/bjdgyc/anylink/dbdata"
	"github.com/bjdgyc/anylink/sessdata"
	gosysctl "github.com/lorenzosaino/go-sysctl"
)

func Start() {
	dbdata.Start()
	sessdata.Start()
	cron.Start()

	admin.InitLockManager() // 初始化防爆破定时器和IP白名单

	// 开启服务器转发
	err := gosysctl.Set("net.ipv4.ip_forward", "1")
	if err != nil {
		base.Warn(err)
	}

	val, err := gosysctl.Get("net.ipv4.ip_forward")
	if val != "1" {
		log.Fatal("Please exec 'sysctl -w net.ipv4.ip_forward=1' ")
	}
	// os.Exit(0)
	// execCmd([]string{"sysctl -w net.ipv4.ip_forward=1"})

	switch base.Cfg.LinkMode {
	case base.LinkModeTUN:
		checkTun()
	case base.LinkModeTAP:
		checkTap()
	case base.LinkModeMacvtap:
		checkMacvtap()
	default:
		base.Fatal("LinkMode is err")
	}

	// 计算profile.xml的hash
	b, err := os.ReadFile(base.Cfg.Profile)
	if err != nil {
		panic(err)
	}
	ha := sha1.Sum(b)
	profileHash = hex.EncodeToString(ha[:])

	go admin.StartAdmin()
	go startTls()
	go startDtls()

	go logAuditBatch()
}

func Stop() {
	_ = dbdata.Stop()
	destroyVtap()
}
