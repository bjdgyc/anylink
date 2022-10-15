package cron

import (
	"time"

	"github.com/bjdgyc/anylink/sessdata"
	"github.com/go-co-op/gocron"
)

func Start() {
	s := gocron.NewScheduler(time.Local)
	s.Cron("0 * * * *").Do(ClearAudit)
	s.Cron("0 * * * *").Do(ClearStatsInfo)
	s.Every("1s").Do(sessdata.CloseUserLimittimeSession)
	s.StartAsync()
}
