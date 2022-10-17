package cron

import (
	"time"

	"github.com/bjdgyc/anylink/base"
	"github.com/bjdgyc/anylink/dbdata"
)

const siLifeDay = 30

// 清除图表数据
func ClearStatsInfo() {
	_, timesUp := isClearTime()
	if !timesUp {
		return
	}
	ts := getTimeAgo(siLifeDay)
	for _, item := range dbdata.StatsInfoIns.Actions {
		affected, err := dbdata.StatsInfoIns.ClearStatsInfo(item, ts)
		base.Info("Cron ClearStatsInfo  "+item+": ", affected, err)
	}
}

// 是否到了"清理时间"
func isClearTime() (int, bool) {
	dataLog, err := dbdata.SettingGetAuditLog()
	if err != nil {
		base.Error("Cron SettingGetLog: ", err)
		return -1, false
	}
	currentTime := time.Now().Format("15:04")
	// 未到"清理时间"时, 则返回
	if dataLog.ClearTime != currentTime {
		return -1, false
	}
	return dataLog.LifeDay, true
}

// 根据存储时长，获取清理日期
func getTimeAgo(days int) string {
	var timeS string
	ts := time.Now().AddDate(0, 0, -days)
	tsZero := time.Date(ts.Year(), ts.Month(), ts.Day(), 0, 0, 0, 0, time.Local)
	timeS = tsZero.Format(dbdata.LayoutTimeFormat)
	// UTC
	switch base.Cfg.DbType {
	case "sqlite3", "postgres":
		timeS = tsZero.UTC().Format(dbdata.LayoutTimeFormat)
	}
	return timeS
}
