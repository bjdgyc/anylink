package handler

import (
	"github.com/bjdgyc/anylink/common"
	"github.com/bjdgyc/anylink/dbdata"
	"github.com/bjdgyc/anylink/sessdata"
)

func Start() {
	dbdata.Start()
	sessdata.Start()

	checkTun()
	if common.ServerCfg.LinkMode == common.LinkModeTAP {
		checkTap()
	}
	go startAdmin()
	go startTls()
	go startDtls()
}

func Stop() {
	dbdata.Stop()
}
