package handler

import (
	"context"
	"encoding/xml"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/bjdgyc/anylink/base"
)

// var lockManager = admin.GetLockManager()

const loginStatusKey = "login_status"

type HttpContext struct {
	LoginStatus bool // 登录状态
}

// 防爆破中间件
func antiBruteForce(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, old_r *http.Request) {
		// 防爆破功能全局开关
		if !base.Cfg.AntiBruteForce {
			next.ServeHTTP(w, old_r)
			return
		}

		// 非并发安全
		hc := &HttpContext{}
		ctx := context.WithValue(context.Background(), loginStatusKey, hc)
		r := old_r.WithContext(ctx)

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		cr := ClientRequest{}
		err = xml.Unmarshal(body, &cr)
		if err != nil {
			http.Error(w, "Failed to parse XML", http.StatusBadRequest)
			return
		}

		username := cr.Auth.Username
		if r.URL.Path == "/otp-verification" {
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
			username = sessionData.ClientRequest.Auth.Username
		}
		ip, _, err := net.SplitHostPort(r.RemoteAddr) // 提取纯 IP 地址，去掉端口号
		if err != nil {
			http.Error(w, "Unable to parse IP address", http.StatusInternalServerError)
			return
		}

		now := time.Now()
		// 检查IP是否在白名单中
		if lockManager.IsWhitelisted(ip) {
			r.Body = io.NopCloser(strings.NewReader(string(body)))
			next.ServeHTTP(w, r)
			return
		}

		// 检查全局 IP 锁定
		if base.Cfg.MaxGlobalIPBanCount > 0 && lockManager.CheckGlobalIPLock(ip, now) {
			base.Warn("IP", ip, "is globally locked. Try again later.")
			http.Error(w, "Account globally locked due to too many failed attempts. Try again later.", http.StatusTooManyRequests)
			return
		}

		// 检查全局用户锁定
		if base.Cfg.MaxGlobalUserBanCount > 0 && lockManager.CheckGlobalUserLock(username, now) {
			base.Warn("User", username, "is globally locked. Try again later.")
			http.Error(w, "Account globally locked due to too many failed attempts. Try again later.", http.StatusTooManyRequests)
			return
		}

		// 检查单个用户的 IP 锁定
		if base.Cfg.MaxBanCount > 0 && lockManager.CheckUserIPLock(username, ip, now) {
			base.Warn("IP", ip, "is locked for user", username, "Try again later.")
			http.Error(w, "Account locked due to too many failed attempts. Try again later.", http.StatusTooManyRequests)
			return
		}

		// 重新设置请求体以便后续处理器可以访问
		r.Body = io.NopCloser(strings.NewReader(string(body)))

		// 调用下一个处理器
		next.ServeHTTP(w, r)

		// 检查登录状态
		// Status, _ := lockManager.LoginStatus.Load(loginStatusKey)
		// loginStatus, _ := Status.(bool)

		loginStatus := hc.LoginStatus

		// 更新用户登录状态
		lockManager.UpdateGlobalIPLock(ip, now, loginStatus)
		lockManager.UpdateGlobalUserLock(username, now, loginStatus)
		lockManager.UpdateUserIPLock(username, ip, now, loginStatus)

		// 清除登录状态
		// lockManager.LoginStatus.Delete(loginStatusKey)
	})
}
