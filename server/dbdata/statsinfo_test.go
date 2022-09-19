package dbdata

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatsInfo(t *testing.T) {
	ast := assert.New(t)

	preIpData()
	defer closeIpdata()

	ast.True(StatsInfoIns.ValidAction("online"))
	ast.False(StatsInfoIns.ValidAction("diskio"))
	ast.True(StatsInfoIns.ValidScope("30d"))
	ast.False(StatsInfoIns.ValidScope("60d"))

	up := uint32(100)
	down := uint32(300)
	upGroups := map[int]uint32{1: up}
	downGroups := map[int]uint32{1: down}
	numGroups := map[int]int{1: 5}
	// online
	numData, _ := json.Marshal(numGroups)
	so := &StatsOnline{Num: 1, NumGroups: string(numData)}
	// network
	upData, _ := json.Marshal(upGroups)
	downData, _ := json.Marshal(downGroups)
	sn := &StatsNetwork{Up: up, Down: down, UpGroups: string(upData), DownGroups: string(downData)}
	// cpu
	sc := &StatsCpu{Percent: 0.3}
	// mem
	sm := &StatsMem{Percent: 24.50}

	StatsInfoIns.SetRealTime("online", so)
	StatsInfoIns.GetRealTime("online")
	StatsInfoIns.SaveStatsInfo(so, sn, sc, sm)

	var err error
	_, err = StatsInfoIns.GetData("online", "1h")
	ast.Nil(err)

	_, err = StatsInfoIns.GetData("network", "1h")
	ast.Nil(err)

	_, err = StatsInfoIns.GetData("cpu", "1h")
	ast.Nil(err)

	_, err = StatsInfoIns.GetData("mem", "1h")
	ast.Nil(err)

}
