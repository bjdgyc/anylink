package dbdata

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/ivpusic/grpool"
	"github.com/spf13/cast"
	"xorm.io/xorm"
)

const (
	UserAuthFail    = 0 // 认证失败
	UserAuthSuccess = 1 // 认证成功
	UserConnected   = 2 // 连线成功
	UserLogout      = 3 // 用户登出
)

type UserActLogProcess struct {
	Pool      *grpool.Pool
	StatusOps []string
	OsOps     []string
	ClientOps []string
	InfoOps   []string
}

var (
	UserActLogIns = &UserActLogProcess{
		Pool: grpool.NewPool(1, 100),
		StatusOps: []string{ // 操作类型
			UserAuthFail:    "认证失败",
			UserAuthSuccess: "认证成功",
			UserConnected:   "连接成功",
			UserLogout:      "用户登出",
		},
		OsOps: []string{ // 操作系统
			0: "Windows",
			1: "macOS",
			2: "Linux",
			3: "Android",
			4: "iOS",
		},
		ClientOps: []string{ // 客户端
			0: "AnyConnect",
			1: "OpenConnect",
			2: "unknown",
		},
		InfoOps: []string{ // 信息
			0: "用户掉线",
			1: "用户/客户端主动断开",
			2: "用户被踢下线(管理员/账号过期)",
		},
	}
)

// 异步写入用户操作日志
func (ua *UserActLogProcess) Add(u UserActLog, userAgent string) {
	os_idx, client_idx, ver := ua.ParseUserAgent(userAgent)
	u.Os = os_idx
	u.Client = client_idx
	u.Version = ver
	u.RemoteAddr = strings.Split(u.RemoteAddr, ":")[0]
	// remove extra characters
	infoSlice := strings.Split(u.Info, " ")
	infoLen := len(infoSlice)
	if infoLen > 1 {
		if u.Username == infoSlice[0] {
			u.Info = strings.Join(infoSlice[1:], " ")
		}
		// delete - char
		if infoLen > 2 && infoSlice[1] == "-" {
			u.Info = u.Info[2:]
		}
	}
	UserActLogIns.Pool.JobQueue <- func() {
		_ = Add(u)
	}
}

// 转义操作类型, 方便vue显示
func (ua *UserActLogProcess) GetStatusOpsWithTag() interface{} {
	type StatusTag struct {
		Key   int    `json:"key"`
		Value string `json:"value"`
		Tag   string `json:"tag"`
	}
	var res []StatusTag
	for k, v := range ua.StatusOps {
		tag := "info"
		switch k {
		case UserAuthFail:
			tag = "danger"
		case UserAuthSuccess:
			tag = "success"
		case UserConnected:
			tag = ""
		}
		res = append(res, StatusTag{k, v, tag})
	}
	return res
}

func (ua *UserActLogProcess) GetInfoOpsById(id uint8) string {
	infoMap := ua.InfoOps
	return infoMap[id]
}

func (ua *UserActLogProcess) ParseUserAgent(userAgent string) (os_idx, client_idx uint8, ver string) {
	// os
	os_idx = 2
	if strings.Contains(userAgent, "windows") {
		os_idx = 0
	} else if strings.Contains(userAgent, "mac os") || strings.Contains(userAgent, "darwin_i386") {
		os_idx = 1
	} else if strings.Contains(userAgent, "darwin_arm") || strings.Contains(userAgent, "apple") {
		os_idx = 4
	} else if strings.Contains(userAgent, "android") {
		os_idx = 3
	}
	// client
	client_idx = 2
	if strings.Contains(userAgent, "anyconnect") {
		client_idx = 0
	} else if strings.Contains(userAgent, "openconnect") {
		client_idx = 1
	}
	// ver
	uaSlice := strings.Split(userAgent, " ")
	ver = uaSlice[len(uaSlice)-1]
	if ver[0] == 'v' {
		ver = ver[1:]
	}
	if !regexp.MustCompile(`^(\d+\.?)+$`).MatchString(ver) {
		ver = ""
	}
	return
}

// 清除用户操作日志
func (ua *UserActLogProcess) ClearUserActLog(ts string) (int64, error) {
	affected, err := xdb.Where("created_at < '" + ts + "'").Delete(&UserActLog{})
	return affected, err
}

// 后台筛选用户操作日志
func (ua *UserActLogProcess) GetSession(values url.Values) *xorm.Session {
	session := xdb.Where("1=1")
	if values.Get("username") != "" {
		session.And("username = ?", values.Get("username"))
	}
	if values.Get("sdate") != "" {
		session.And("created_at >= ?", values.Get("sdate")+" 00:00:00'")
	}
	if values.Get("edate") != "" {
		session.And("created_at <= ?", values.Get("edate")+" 23:59:59'")
	}
	if values.Get("status") != "" {
		session.And("status = ?", cast.ToUint8(values.Get("status"))-1)
	}
	if values.Get("os") != "" {
		session.And("os = ?", cast.ToUint8(values.Get("os"))-1)
	}
	if values.Get("sort") == "1" {
		session.OrderBy("id desc")
	} else {
		session.OrderBy("id asc")
	}
	return session
}
