package sessdata

import (
	"encoding/json"
	"time"

	"github.com/bjdgyc/anylink/dbdata"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

const (
	StatsCycleSec = 10 // 统计周期（秒）
	AddCycleSec   = 60 // 记录到数据表周期（秒）
)

func saveStatsInfo() {
	go func() {
		tick := time.NewTicker(time.Second * StatsCycleSec)
		count := 0
		for range tick.C {
			up := uint32(0)
			down := uint32(0)
			upGroups := make(map[int]uint32)
			downGroups := make(map[int]uint32)
			numGroups := make(map[int]int)
			onlineNum := 0
			sessMux.Lock()
			for _, v := range sessions {
				v.mux.Lock()
				if v.IsActive {
					// 在线人数
					onlineNum += 1
					numGroups[v.CSess.Group.Id] += 1
					// 网络吞吐
					userUp := v.CSess.BandwidthUpPeriod.Load()
					userDown := v.CSess.BandwidthDownPeriod.Load()
					if userUp > 0 {
						upGroups[v.CSess.Group.Id] += userUp
					}
					if userDown > 0 {
						downGroups[v.CSess.Group.Id] += userDown
					}
					up += userUp
					down += userDown
				}
				v.mux.Unlock()
			}
			sessMux.Unlock()

			tNow := time.Now()
			// online
			numData, _ := json.Marshal(numGroups)
			so := dbdata.StatsOnline{Num: onlineNum, NumGroups: string(numData), CreatedAt: tNow}
			// network
			upData, _ := json.Marshal(upGroups)
			downData, _ := json.Marshal(downGroups)
			sn := dbdata.StatsNetwork{Up: up, Down: down, UpGroups: string(upData), DownGroups: string(downData), CreatedAt: tNow}
			// cpu
			sc := dbdata.StatsCpu{Percent: getCpuPercent(), CreatedAt: tNow}
			// mem
			sm := dbdata.StatsMem{Percent: getMemPercent(), CreatedAt: tNow}
			count++
			// 是否保存至数据库
			save := count*StatsCycleSec >= AddCycleSec
			// 历史数据
			if save {
				count = 0
			}
			// 设置统计数据
			setStatsData(save, so, sn, sc, sm)
		}
	}()
}

func setStatsData(save bool, so dbdata.StatsOnline, sn dbdata.StatsNetwork, sc dbdata.StatsCpu, sm dbdata.StatsMem) {
	// 实时数据
	dbdata.StatsInfoIns.SetRealTime("online", so)
	dbdata.StatsInfoIns.SetRealTime("network", sn)
	dbdata.StatsInfoIns.SetRealTime("cpu", sc)
	dbdata.StatsInfoIns.SetRealTime("mem", sm)
	if !save {
		return
	}
	dbdata.StatsInfoIns.SaveStatsInfo(so, sn, sc, sm)
}

func getCpuPercent() float64 {
	cpuUsedPercent, _ := cpu.Percent(0, false)
	percent := cpuUsedPercent[0]
	if percent == 0 {
		percent = 1
	}
	return decimal(percent)
}

func getMemPercent() float64 {
	m, _ := mem.VirtualMemory()
	return decimal(m.UsedPercent)
}

func decimal(f float64) float64 {
	i := int(f * 100)
	return float64(i) / 100
}
