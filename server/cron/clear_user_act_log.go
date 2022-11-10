package cron

import (
	"github.com/bjdgyc/anylink/base"
	"github.com/bjdgyc/anylink/dbdata"
)

// 清除用户活动日志
func ClearUserActLog() {
	lifeDay, timesUp := isClearTime()
	if !timesUp {
		return
	}
	// 当审计日志永久保存时，则退出
	if lifeDay <= 0 {
		return
	}
	affected, err := dbdata.UserActLogIns.ClearUserActLog(getTimeAgo(lifeDay))
	base.Info("Cron ClearUserActLog: ", affected, err)
}
