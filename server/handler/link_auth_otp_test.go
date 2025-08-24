package handler

import (
	"bytes"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/bjdgyc/anylink/base"
	"github.com/bjdgyc/anylink/dbdata"
	"github.com/stretchr/testify/assert"
	"github.com/xlzd/gotp"
)

func TestSessionStore(t *testing.T) {
	ast := assert.New(t)

	// 测试会话存储基本功能
	store := NewSessionStore()
	sessionID := "test-session-123"

	// 创建测试会话数据
	authSession := &AuthSession{
		ClientRequest: &ClientRequest{
			Auth: auth{
				Username:  "test-user",
				OtpSecret: "JBSWY3DPEHPK3PXP",
			},
			GroupSelect: "test-group",
		},
		UserActLog: &dbdata.UserActLog{
			Username: "test-user",
			Status:   dbdata.UserAuthSuccess,
		},
	}

	// 测试保存会话
	store.SaveAuthSession(sessionID, authSession)

	// 测试获取会话
	retrievedSession, err := store.GetAuthSession(sessionID)
	ast.Nil(err)
	ast.NotNil(retrievedSession)
	ast.Equal("test-user", retrievedSession.ClientRequest.Auth.Username)

	// 测试获取不存在的会话
	_, err = store.GetAuthSession("nonexistent-session")
	ast.NotNil(err)
	ast.Contains(err.Error(), "auth session not found")

	// 测试删除会话
	store.DeleteAuthSession(sessionID)
	_, err = store.GetAuthSession(sessionID)
	ast.NotNil(err)
}

func TestGenerateSessionID(t *testing.T) {
	ast := assert.New(t)

	// 测试会话ID生成
	sessionID, err := GenerateSessionID()
	ast.Nil(err)
	ast.NotEmpty(sessionID)
	ast.Equal(32, len(sessionID))

	// 测试生成的ID唯一性
	sessionID2, err := GenerateSessionID()
	ast.Nil(err)
	ast.NotEqual(sessionID, sessionID2)
}

func TestCookieOperations(t *testing.T) {
	ast := assert.New(t)

	// 测试设置和获取Cookie
	w := httptest.NewRecorder()
	SetCookie(w, "test-cookie", "test-value", 3600)

	cookies := w.Result().Cookies()
	ast.Equal(1, len(cookies))
	ast.Equal("test-cookie", cookies[0].Name)
	ast.Equal("test-value", cookies[0].Value)
	ast.True(cookies[0].HttpOnly)
	ast.True(cookies[0].Secure)

	// 测试从请求中获取Cookie
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(cookies[0])

	value, err := GetCookie(req, "test-cookie")
	ast.Nil(err)
	ast.Equal("test-value", value)

	// 测试获取不存在的Cookie
	_, err = GetCookie(req, "nonexistent-cookie")
	ast.NotNil(err)

	// 测试删除Cookie
	w2 := httptest.NewRecorder()
	DeleteCookie(w2, "test-cookie")
	deleteCookies := w2.Result().Cookies()
	ast.Equal(1, len(deleteCookies))
	ast.Equal("test-cookie", deleteCookies[0].Name)
	ast.Equal("", deleteCookies[0].Value)
	ast.Equal(-1, deleteCookies[0].MaxAge)
}

func TestLinkAuthOtp(t *testing.T) {
	base.Test()
	ast := assert.New(t)

	base.Cfg.DisplayError = true

	// 设置测试数据库
	preIpData()
	defer closeIpdata()

	// 创建测试组
	group := "otp-test-group"
	dns := []dbdata.ValData{{Val: "8.8.8.8"}}
	g := dbdata.Group{Name: group, Status: 1, ClientDns: dns}
	err := dbdata.SetGroup(&g)
	ast.Nil(err)

	// 创建测试用户
	username := "otp-test-user"
	otpSecret := "JBSWY3DPEHPK3PXP"
	u := dbdata.User{
		Username:  username,
		Groups:    []string{group},
		Status:    1,
		OtpSecret: otpSecret,
	}
	err = dbdata.SetUser(&u)
	ast.Nil(err)

	// 生成有效的OTP代码
	totp := gotp.NewDefaultTOTP(otpSecret)
	validOtp := totp.Now()

	// 创建测试会话
	sessionID := "test-otp-session"
	authSession := &AuthSession{
		ClientRequest: &ClientRequest{
			Auth: auth{
				Username:  username,
				OtpSecret: otpSecret,
			},
			GroupSelect: group,
			UserAgent:   "test-agent",
		},
		UserActLog: &dbdata.UserActLog{
			Username: username,
			Status:   dbdata.UserAuthSuccess,
		},
	}
	SessStore.SaveAuthSession(sessionID, authSession)

	// 测试成功的OTP验证
	t.Run("ValidOTP", func(t *testing.T) {
		ast := assert.New(t)

		// 创建OTP验证请求
		clientReq := ClientRequest{
			Auth: auth{
				SecondaryPassword: validOtp,
			},
		}
		reqBody, _ := xml.Marshal(clientReq)

		req := httptest.NewRequest("POST", "/otp-verification", bytes.NewReader(reqBody))
		req.AddCookie(&http.Cookie{Name: "auth-session-id", Value: sessionID})
		w := httptest.NewRecorder()

		LinkAuth_otp(w, req)

		ast.Equal(http.StatusOK, w.Code)
		// 验证会话已被删除
		_, err := SessStore.GetAuthSession(sessionID)
		ast.NotNil(err)
	})

	// 测试无效的OTP代码
	t.Run("InvalidOTP", func(t *testing.T) {
		ast := assert.New(t)

		// 重新创建会话（因为上一个测试中被删除了）
		SessStore.SaveAuthSession(sessionID+"2", authSession)

		clientReq := ClientRequest{
			Auth: auth{
				SecondaryPassword: "123456", // 无效的OTP
			},
		}
		reqBody, _ := xml.Marshal(clientReq)

		req := httptest.NewRequest("POST", "/otp-verification", bytes.NewReader(reqBody))
		req.AddCookie(&http.Cookie{Name: "auth-session-id", Value: sessionID + "2"})
		w := httptest.NewRecorder()

		LinkAuth_otp(w, req)

		ast.Equal(http.StatusOK, w.Code)
		// 验证响应包含错误信息
		ast.Contains(w.Body.String(), "OTP 动态码错误")
	})

	// 测试无效会话
	t.Run("InvalidSession", func(t *testing.T) {
		ast := assert.New(t)

		clientReq := ClientRequest{
			Auth: auth{
				SecondaryPassword: validOtp,
			},
		}
		reqBody, _ := xml.Marshal(clientReq)

		req := httptest.NewRequest("POST", "/otp-verification", bytes.NewReader(reqBody))
		req.AddCookie(&http.Cookie{Name: "auth-session-id", Value: "invalid-session"})
		w := httptest.NewRecorder()

		LinkAuth_otp(w, req)

		ast.Equal(http.StatusUnauthorized, w.Code)
	})

	// 测试缺少会话Cookie
	t.Run("MissingSessionCookie", func(t *testing.T) {
		ast := assert.New(t)

		clientReq := ClientRequest{
			Auth: auth{
				SecondaryPassword: validOtp,
			},
		}
		reqBody, _ := xml.Marshal(clientReq)

		req := httptest.NewRequest("POST", "/otp-verification", bytes.NewReader(reqBody))
		w := httptest.NewRecorder()

		LinkAuth_otp(w, req)

		ast.Equal(http.StatusUnauthorized, w.Code)
	})
}

func TestCreateSession(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("在GitHub Actions中跳过此测试")
		return
	}
	base.Test()
	ast := assert.New(t)

	preIpData()
	defer closeIpdata()

	base.Cfg.EnableBanner = true

	other := &dbdata.SettingOther{Banner: "测试横幅内容"}
	err := dbdata.SettingSet(other)
	ast.Nil(err)

	// 创建测试数据
	group := "session-test-group"
	username := "session-test-user"

	dns := []dbdata.ValData{{Val: "8.8.8.8"}}
	g := dbdata.Group{Name: group, Status: 1, ClientDns: dns}
	err = dbdata.SetGroup(&g)
	ast.Nil(err)

	u := dbdata.User{Username: username, Groups: []string{group}, Status: 1}
	err = dbdata.SetUser(&u)
	ast.Nil(err)

	// 创建认证会话数据
	authSession := &AuthSession{
		ClientRequest: &ClientRequest{
			Auth: auth{
				Username: username,
			},
			GroupSelect: group,
			UserAgent:   "test-agent",
			DeviceId: deviceId{
				UniqueIdGlobal: "test-device-id",
			},
			MacAddressList: macAddressList{
				MacAddress: "00:11:22:33:44:55",
			},
			RemoteAddr: "192.168.1.100",
		},
		UserActLog: &dbdata.UserActLog{
			Username:        username,
			Status:          dbdata.UserAuthSuccess,
			DeviceType:      "test-device",
			PlatformVersion: "test-platform",
		},
	}

	// 测试会话创建
	req := httptest.NewRequest("POST", "/", nil)
	req.RemoteAddr = "192.168.1.100:12345"
	w := httptest.NewRecorder()

	CreateSession(w, req, authSession)

	ast.Equal(http.StatusOK, w.Code)
	// 验证响应包含会话信息
	ast.Contains(w.Body.String(), "session-token")
	ast.Contains(w.Body.String(), "测试横幅内容")

	base.Cfg.EnableBanner = false

	w2 := httptest.NewRecorder()
	CreateSession(w2, req, authSession)

	ast.Equal(http.StatusOK, w2.Code)
	ast.NotContains(w2.Body.String(), "测试横幅内容")
}

func preIpData() {
	// 设置测试模式
	base.Test()

	// 创建临时数据库文件
	tmpDb := path.Join(os.TempDir(), "anylink_otp_test.db")

	// 设置数据库配置
	base.Cfg.DbType = "sqlite3"
	base.Cfg.DbSource = tmpDb

	// 设置其他必要的配置
	base.Cfg.Ipv4CIDR = "192.168.3.0/24"
	base.Cfg.Ipv4Gateway = "192.168.3.1"
	base.Cfg.Ipv4Start = "192.168.3.100"
	base.Cfg.Ipv4End = "192.168.3.150"
	base.Cfg.MaxClient = 100
	base.Cfg.MaxUserClient = 3
	base.Cfg.IpLease = 5

	// 启动数据库
	dbdata.Start()
}

func closeIpdata() {
	_ = dbdata.Stop()
	tmpDb := path.Join(os.TempDir(), "anylink_otp_test.db")
	os.Remove(tmpDb)
}
