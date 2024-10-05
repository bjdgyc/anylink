package handler

import (
	"crypto/sha1"
	"encoding/hex"
	"os"

	"github.com/bjdgyc/anylink/admin"
	"github.com/bjdgyc/anylink/base"
	"github.com/bjdgyc/anylink/cron"
	"github.com/bjdgyc/anylink/dbdata"
	"github.com/bjdgyc/anylink/sessdata"
)

func Start() {
	dbdata.Start()
	sessdata.Start()
	cron.Start()

	initAntiBruteForce() //初始化防爆破定时器和IP白名单

	// 开启服务器转发
	err := execCmd([]string{"sysctl -w net.ipv4.ip_forward=1"})
	if err != nil {
		base.Fatal(err)
	}

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
