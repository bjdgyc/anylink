package handler

import (
	"crypto/md5"
	"encoding/xml"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"

	"github.com/bjdgyc/anylink/admin"
	"github.com/bjdgyc/anylink/base"
	"github.com/bjdgyc/anylink/dbdata"
	"github.com/bjdgyc/anylink/pkg/utils"
	"github.com/bjdgyc/anylink/sessdata"
)

var SessStore = NewSessionStore()
var lockManager = admin.GetLockManager()

// const maxOtpErrCount = 3

type AuthSession struct {
	ClientRequest *ClientRequest
	UserActLog    *dbdata.UserActLog
	// OtpErrCount   atomic.Uint32 // otp错误次数
}

// 存储临时会话信息
type SessionStore struct {
	session map[string]*AuthSession
	mu      sync.Mutex
}

func NewSessionStore() *SessionStore {
	return &SessionStore{
		session: make(map[string]*AuthSession),
	}
}

func (s *SessionStore) SaveAuthSession(sessionID string, session *AuthSession) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.session[sessionID] = session
}

func (s *SessionStore) GetAuthSession(sessionID string) (*AuthSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.session[sessionID]
	if !exists {
		return nil, fmt.Errorf("auth session not found")
	}

	return session, nil
}

func (s *SessionStore) DeleteAuthSession(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.session, sessionID)
}

// func (a *AuthSession) AddOtpErrCount(i int) int {
// 	newI := a.OtpErrCount.Add(uint32(i))
// 	return int(newI)
// }

func GenerateSessionID() (string, error) {
	sessionID := utils.RandomRunes(32)
	if sessionID == "" {
		return "", fmt.Errorf("failed to generate session ID")
	}

	return sessionID, nil
}

func SetCookie(w http.ResponseWriter, name, value string, maxAge int) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   maxAge,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)
}

func GetCookie(r *http.Request, name string) (string, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", fmt.Errorf("failed to get cookie: %v", err)
	}
	return cookie.Value, nil
}

func DeleteCookie(w http.ResponseWriter, name string) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    "",
		MaxAge:   -1,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, cookie)
}
func CreateSession(w http.ResponseWriter, r *http.Request, authSession *AuthSession) {
	cr := authSession.ClientRequest
	ua := authSession.UserActLog

	lockManager.UpdateLoginStatus(cr.Auth.Username, r.RemoteAddr, true) // 更新登录成功状态

	sess := sessdata.NewSession("")
	sess.Username = cr.Auth.Username
	sess.Group = cr.GroupSelect
	oriMac := cr.MacAddressList.MacAddress
	sess.UniqueIdGlobal = cr.DeviceId.UniqueIdGlobal
	sess.UserAgent = cr.UserAgent
	sess.DeviceType = ua.DeviceType
	sess.PlatformVersion = ua.PlatformVersion
	sess.RemoteAddr = cr.RemoteAddr
	// 获取客户端mac地址
	sess.UniqueMac = true
	macHw, err := net.ParseMAC(oriMac)
	if err != nil {
		var sum [16]byte
		if sess.UniqueIdGlobal != "" {
			sum = md5.Sum([]byte(sess.UniqueIdGlobal))
		} else {
			sum = md5.Sum([]byte(sess.Token))
			sess.UniqueMac = false
		}
		macHw = sum[0:5] // 5个byte
		macHw = append([]byte{0x02}, macHw...)
		sess.MacAddr = macHw.String()
	}
	sess.MacHw = macHw
	// 统一macAddr的格式
	sess.MacAddr = macHw.String()

	other := &dbdata.SettingOther{}
	dbdata.SettingGet(other)
	rd := RequestData{
		SessionId:    sess.Sid,
		SessionToken: sess.Sid + "@" + sess.Token,
		// Banner:       other.Banner,
		ProfileName: base.Cfg.ProfileName,
		ProfileHash: profileHash,
		CertHash:    certHash,
	}
	if base.Cfg.EnableBanner {
		rd.Banner = other.Banner
	}

	w.WriteHeader(http.StatusOK)
	tplRequest(tpl_complete, w, rd)
	base.Info("login", cr.Auth.Username, cr.UserAgent)
}

func LinkAuth_otp(w http.ResponseWriter, r *http.Request) {
	sessionID, err := GetCookie(r, "auth-session-id")
	if err != nil {
		http.Error(w, "Invalid session, please login again", http.StatusUnauthorized)
		return
	}

	sessionData, err := SessStore.GetAuthSession(sessionID)
	if err != nil {
		http.Error(w, "Invalid session, please login again", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		base.Error(err)
		SessStore.DeleteAuthSession(sessionID)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	cr := ClientRequest{}
	err = xml.Unmarshal(body, &cr)
	if err != nil {
		base.Error(err)
		SessStore.DeleteAuthSession(sessionID)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ua := sessionData.UserActLog
	username := sessionData.ClientRequest.Auth.Username
	otpSecret := sessionData.ClientRequest.Auth.OtpSecret
	otp := cr.Auth.SecondaryPassword

	// 锁定状态判断
	if !lockManager.CheckLocked(username, r.RemoteAddr) {
		w.WriteHeader(http.StatusTooManyRequests)
		SessStore.DeleteAuthSession(sessionID)
		return
	}

	// 动态码错误
	if !dbdata.CheckOtp(username, otp, otpSecret) {
		lockManager.UpdateLoginStatus(username, r.RemoteAddr, false) // 记录登录失败状态

		base.Warn("OTP 动态码错误", username, r.RemoteAddr)
		ua.Info = "OTP 动态码错误"
		ua.Status = dbdata.UserAuthFail
		dbdata.UserActLogIns.Add(*ua, sessionData.ClientRequest.UserAgent)

		w.WriteHeader(http.StatusOK)
		data := RequestData{Error: "请求错误"}
		if base.Cfg.DisplayError {
			data.Error = "OTP 动态码错误"
		}
		tplRequest(tpl_otp, w, data)
		return
	}
	CreateSession(w, r, sessionData)

	// 删除临时会话信息
	SessStore.DeleteAuthSession(sessionID)
	// DeleteCookie(w, "auth-session-id")
}

var auth_otp = `<?xml version="1.0" encoding="UTF-8"?>
<config-auth client="vpn" type="auth-request" aggregate-auth-version="2">
    <auth id="otp-verification">
        <title>OTP 动态码验证</title>
        <message>请输入您的 OTP 动态码</message>
        {{if .Error}}
        <error id="otp-verification" param1="{{.Error}}" param2="">验证失败:  %s</error>
        {{end}}		
        <form method="post" action="/otp-verification">
            <input type="password" name="secondary_password" label="OTPCode:"/>
        </form>
    </auth>
</config-auth>`
