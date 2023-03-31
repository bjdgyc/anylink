package cron

import (
	"time"

	"github.com/bjdgyc/anylink/admin"
	"github.com/bjdgyc/anylink/sessdata"
	"github.com/go-co-op/gocron"
)

func Start() {
	s := gocron.NewScheduler(time.Local)
	s.Cron("0 * * * *").Do(ClearAudit)
	s.Cron("0 * * * *").Do(ClearStatsInfo)
	s.Cron("0 * * * *").Do(ClearUserActLog)
	s.Every(1).Day().At("00:00").Do(sessdata.CloseUserLimittimeSession)
	s.Every(1).Day().At("00:00").Do(admin.ReNewCert)
	s.StartAsync()
}
