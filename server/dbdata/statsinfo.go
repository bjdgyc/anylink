package dbdata

import (
	"container/list"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/bjdgyc/anylink/base"
)

const (
	LayoutTimeFormat    = "2006-01-02 15:04:05"
	LayoutTimeFormatMin = "2006-01-02 15:04"
	RealTimeMaxSize     = 120 // 实时数据最大保存条数
)

type StatsInfo struct {
	RealtimeData map[string]*list.List
	Actions      []string
	Scopes       []string
}

type ScopeDetail struct {
	sTime   time.Time
	eTime   time.Time
	minutes int
	fsTime  string
	feTime  string
}

var StatsInfoIns *StatsInfo

func init() {
	StatsInfoIns = &StatsInfo{
		Actions:      []string{"online", "network", "cpu", "mem"},
		Scopes:       []string{"rt", "1h", "24h", "3d", "7d", "30d"},
		RealtimeData: make(map[string]*list.List),
	}
	for _, v := range StatsInfoIns.Actions {
		StatsInfoIns.RealtimeData[v] = list.New()
	}
}

// 校验统计类型值
func (s *StatsInfo) ValidAction(action string) bool {
	for _, item := range s.Actions {
		if item == action {
			return true
		}
	}
	return false
}

// 校验日期范围值
func (s *StatsInfo) ValidScope(scope string) bool {
	for _, item := range s.Scopes {
		if item == scope {
			return true
		}
	}
	return false
}

// 设置实时统计数据
func (s *StatsInfo) SetRealTime(action string, val interface{}) {
	if s.RealtimeData[action].Len() >= RealTimeMaxSize {
		ele := s.RealtimeData[action].Front()
		s.RealtimeData[action].Remove(ele)
	}
	s.RealtimeData[action].PushBack(val)
}

// 获取实时统计数据
func (s *StatsInfo) GetRealTime(action string) (res []interface{}) {
	for e := s.RealtimeData[action].Front(); e != nil; e = e.Next() {
		res = append(res, e.Value)
	}
	return
}

// 保存数据至数据库
func (s *StatsInfo) SaveStatsInfo(so *StatsOnline, sn *StatsNetwork, sc *StatsCpu, sm *StatsMem) {
	if so.Num != 0 {
		_ = Add(so)
	}
	if sn.Up != 0 || sn.Down != 0 {
		_ = Add(sn)
	}
	if sc.Percent != 0 {
		_ = Add(sc)
	}
	if sm.Percent != 0 {
		_ = Add(sm)
	}
}

// 获取统计数据
func (s *StatsInfo) GetData(action string, scope string) (res []interface{}, err error) {
	if scope == "rt" {
		return s.GetRealTime(action), nil
	}
	statsMaps := make(map[string]interface{})
	currSec := fmt.Sprintf("%02d", time.Now().Second())

	// 获取时间段数据
	sd := s.getScopeDetail(scope)
	timeList := s.getTimeList(sd)
	res = make([]interface{}, len(timeList))

	// 获取数据库查询条件
	where := s.getStatsWhere(sd)
	if where == "" {
		return nil, errors.New("不支持的数据库类型: " + base.Cfg.DbType)
	}
	// 查询数据表
	switch action {
	case "online":
		statsRes := []StatsOnline{}
		FindWhere(&statsRes, 0, 0, where, sd.fsTime, sd.feTime)
		for _, v := range statsRes {
			t := v.CreatedAt.Format(LayoutTimeFormatMin)
			statsMaps[t] = v
		}
	case "network":
		statsRes := []StatsNetwork{}
		FindWhere(&statsRes, 0, 0, where, sd.fsTime, sd.feTime)
		for _, v := range statsRes {
			t := v.CreatedAt.Format(LayoutTimeFormatMin)
			statsMaps[t] = v
		}
	case "cpu":
		statsRes := []StatsCpu{}
		FindWhere(&statsRes, 0, 0, where, sd.fsTime, sd.feTime)
		for _, v := range statsRes {
			t := v.CreatedAt.Format(LayoutTimeFormatMin)
			statsMaps[t] = v
		}
	case "mem":
		statsRes := []StatsMem{}
		FindWhere(&statsRes, 0, 0, where, sd.fsTime, sd.feTime)
		for _, v := range statsRes {
			t := v.CreatedAt.Format(LayoutTimeFormatMin)
			statsMaps[t] = v
		}
	}
	// 整合数据
	for i, v := range timeList {
		if mv, ok := statsMaps[v]; ok {
			res[i] = mv
			continue
		}
		t, _ := time.ParseInLocation(LayoutTimeFormat, v+":"+currSec, time.Local)
		switch action {
		case "online":
			res[i] = StatsOnline{CreatedAt: t}
		case "network":
			res[i] = StatsNetwork{CreatedAt: t}
		case "cpu":
			res[i] = StatsCpu{CreatedAt: t}
		case "mem":
			res[i] = StatsMem{CreatedAt: t}
		}
	}
	return
}

// 获取日期范围的明细值
func (s *StatsInfo) getScopeDetail(scope string) (sd *ScopeDetail) {
	sd = &ScopeDetail{}
	t := time.Now()
	sd.eTime = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 59, t.Nanosecond(), time.Local)
	sd.minutes = 0
	switch scope {
	case "1h":
		sd.sTime = sd.eTime.Add(-time.Minute * 60)
		sd.minutes = 1
	case "24h":
		sd.sTime = sd.eTime.AddDate(0, 0, -1)
		sd.minutes = 5
	case "7d":
		sd.sTime = sd.eTime.AddDate(0, 0, -7)
		sd.minutes = 30
	case "30d":
		sd.sTime = sd.eTime.AddDate(0, 0, -30)
		sd.minutes = 150
	}
	if sd.minutes != 0 {
		sd.sTime = sd.sTime.Add(-time.Minute * time.Duration(sd.minutes))
	}
	sd.fsTime = sd.sTime.Format(LayoutTimeFormat)
	sd.feTime = sd.eTime.Format(LayoutTimeFormat)
	// UTC
	switch base.Cfg.DbType {
	case "sqlite3", "postgres":
		sd.fsTime = sd.sTime.UTC().Format(LayoutTimeFormat)
		sd.feTime = sd.eTime.UTC().Format(LayoutTimeFormat)
	}
	return
}

// 针对日期范围进行拆解
func (s *StatsInfo) getTimeList(sd *ScopeDetail) []string {
	subSec := int64(60 * sd.minutes)
	count := (sd.eTime.Unix()-sd.sTime.Unix())/subSec - 1
	eTime := sd.eTime.Unix() - subSec
	timeLists := make([]string, count)
	for i := count - 1; i >= 0; i-- {
		timeLists[i] = time.Unix(eTime, 0).Format(LayoutTimeFormatMin)
		eTime = eTime - subSec
	}
	return timeLists
}

// 获取where条件
func (s *StatsInfo) getStatsWhere(sd *ScopeDetail) (where string) {
	where = "created_at BETWEEN ? AND ?"
	min := strconv.Itoa(sd.minutes)
	switch base.Cfg.DbType {
	case "mysql":
		where += " AND floor(TIMESTAMPDIFF(SECOND, created_at, '" + sd.feTime + "') / 60) % " + min + " = 0"
	case "sqlite3":
		where += " AND CAST(ROUND((JULIANDAY('" + sd.feTime + "') - JULIANDAY(created_at)) * 86400) / 60 as integer) % " + min + " = 0"
	case "postgres":
		where += " AND floor((EXTRACT(EPOCH FROM " + sd.feTime + ") - EXTRACT(EPOCH FROM created_at)) / 60) % " + min + " = 0"
	default:
		where = ""
	}
	return
}
