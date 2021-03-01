package admin

import (
	"fmt"
	"net/http"
	"runtime"

	"github.com/bjdgyc/anylink/dbdata"
	"github.com/bjdgyc/anylink/sessdata"

	"github.com/bjdgyc/anylink/base"
	"github.com/bjdgyc/anylink/pkg/utils"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
)

func SetHome(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]interface{})

	sess := sessdata.OnlineSess()

	data["counts"] = map[string]int{
		"online": len(sess),
		"user":   dbdata.CountAll(&dbdata.User{}),
		"group":  dbdata.CountAll(&dbdata.Group{}),
		"ip_map": dbdata.CountAll(&dbdata.IpMap{}),
	}

	RespSucess(w, data)
}

func SetSystem(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]interface{})

	m, _ := mem.VirtualMemory()
	data["mem"] = map[string]interface{}{
		"total":   utils.HumanByte(m.Total),
		"free":    utils.HumanByte(m.Free),
		"percent": decimal(m.UsedPercent),
	}

	d, _ := disk.Usage("/")
	data["disk"] = map[string]interface{}{
		"total":   utils.HumanByte(d.Total),
		"free":    utils.HumanByte(d.Free),
		"percent": decimal(d.UsedPercent),
	}

	cc, _ := cpu.Counts(true)
	c, _ := cpu.Info()
	ci := c[0]
	cpuUsedPercent, _ := cpu.Percent(0, false)
	cup := cpuUsedPercent[0]
	if cup == 0 {
		cup = 1
	}
	data["cpu"] = map[string]interface{}{
		"core":      cc,
		"modelName": ci.ModelName,
		"ghz":       fmt.Sprintf("%.2f GHz", ci.Mhz/1000),
		"percent":   decimal(cup),
	}

	hi, _ := host.Info()
	l, _ := load.Avg()
	data["sys"] = map[string]interface{}{
		"goOs":      runtime.GOOS,
		"goArch":    runtime.GOARCH,
		"goVersion": runtime.Version(),
		"goroutine": runtime.NumGoroutine(),

		"hostname": hi.Hostname,
		"platform": fmt.Sprintf("%v %v %v", hi.Platform, hi.PlatformFamily, hi.PlatformVersion),
		"kernel":   hi.KernelVersion,

		"load": fmt.Sprint(l.Load1, l.Load5, l.Load15),
	}

	RespSucess(w, data)
}

func SetSoft(w http.ResponseWriter, r *http.Request) {
	data := base.ServerCfg2Slice()
	RespSucess(w, data)
}

func decimal(f float64) float64 {
	i := int(f * 100)
	return float64(i) / 100
}
