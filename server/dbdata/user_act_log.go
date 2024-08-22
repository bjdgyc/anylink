package dbdata

import (
	"net"
	"net/url"
	"regexp"
	"strings"

	"github.com/bjdgyc/anylink/base"
	"github.com/ivpusic/grpool"
	"github.com/spf13/cast"
	"xorm.io/xorm"
)

const (
	UserAuthFail       = 0 // 认证失败
	UserAuthSuccess    = 1 // 认证成功
	UserConnected      = 2 // 连线成功
	UserLogout         = 3 // 用户登出
	UserLogoutLose     = 0 // 用户掉线
	UserLogoutBanner   = 1 // 用户banner弹窗取消
	UserLogoutClient   = 2 // 用户主动登出
	UserLogoutTimeout  = 3 // 用户超时登出
	UserLogoutAdmin    = 4 // 账号被管理员踢下线
	UserLogoutExpire   = 5 // 账号过期被踢下线
	UserIdleTimeout    = 6 // 用户空闲链接超时
	UserLogoutOneAdmin = 7 // 账号被管理员一键下线

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
			0: "Unknown",
			1: "Windows",
			2: "macOS",
			3: "Linux",
			4: "Android",
			5: "iOS",
		},
		ClientOps: []string{ // 客户端
			0: "Unknown",
			1: "AnyConnect",
			2: "OpenConnect",
			3: "AnyLink",
		},
		InfoOps: []string{ // 信息
			UserLogoutLose:     "用户掉线",
			UserLogoutBanner:   "用户取消弹窗/客户端发起的logout",
			UserLogoutClient:   "用户/客户端主动断开",
			UserLogoutTimeout:  "Session过期被踢下线",
			UserLogoutAdmin:    "账号被管理员踢下线",
			UserLogoutExpire:   "账号过期被踢下线",
			UserIdleTimeout:    "用户空闲链接超时",
			UserLogoutOneAdmin: "账号被管理员一键下线",
		},
	}
)

// 异步写入用户操作日志
func (ua *UserActLogProcess) Add(u UserActLog, userAgent string) {
	// os, client, ver
	os_idx, client_idx, ver := ua.ParseUserAgent(userAgent)
	u.Os = os_idx
	u.Client = client_idx
	u.Version = ver
	// u.RemoteAddr = strings.Split(u.RemoteAddr, ":")[0]
	u.RemoteAddr, _, _ = net.SplitHostPort(u.RemoteAddr)
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
	// limit the max length of char
	u.Version = substr(u.Version, 0, 15)
	u.DeviceType = substr(u.DeviceType, 0, 128)
	u.PlatformVersion = substr(u.PlatformVersion, 0, 128)
	u.Info = substr(u.Info, 0, 255)

	UserActLogIns.Pool.JobQueue <- func() {
		err := Add(u)
		if err != nil {
			base.Error("Add UserActLog error: ", err)
		}
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
	if int(id) >= len(ua.InfoOps) {
		return "未知的信息类型"
	}
	return ua.InfoOps[id]
}

// 解析user agent
func (ua *UserActLogProcess) ParseUserAgent(userAgent string) (os_idx, client_idx uint8, ver string) {
	// Unknown
	if len(userAgent) == 0 {
		return 0, 0, ""
	}
	// OS
	os_idx = 0
	if strings.Contains(userAgent, "windows") {
		os_idx = 1
	} else if strings.Contains(userAgent, "mac os") || strings.Contains(userAgent, "darwin_i386") || strings.Contains(userAgent, "darwin_amd64") || strings.Contains(userAgent, "darwin_arm64") {
		os_idx = 2
	} else if strings.Contains(userAgent, "darwin_arm") || strings.Contains(userAgent, "apple") {
		os_idx = 5
	} else if strings.Contains(userAgent, "android") {
		os_idx = 4
	} else if strings.Contains(userAgent, "linux") {
		os_idx = 3
	}
	// Client
	client_idx = 0
	if strings.Contains(userAgent, "anyconnect") {
		client_idx = 1
	} else if strings.Contains(userAgent, "openconnect") {
		client_idx = 2
	} else if strings.Contains(userAgent, "anylink") {
		client_idx = 3
	}
	// Version
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

// 截取字符串
func substr(s string, pos, length int) string {
	runes := []rune(s)
	l := pos + length
	if l > len(runes) {
		l = len(runes)
	}
	return string(runes[pos:l])
}
