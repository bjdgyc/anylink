package handler

import (
	"github.com/bjdgyc/anylink/admin"
	"github.com/bjdgyc/anylink/base"
	"github.com/bjdgyc/anylink/dbdata"
	"github.com/bjdgyc/anylink/sessdata"
)

func Start() {
	dbdata.Start()
	sessdata.Start()
	initBack()
	checkTun()
	if base.Cfg.LinkMode == base.LinkModeTAP {
		checkTap()
	}
	go admin.StartAdmin()
	go startTls()
	go startDtls()
}

func Stop() {
	_ = dbdata.Stop()
}
