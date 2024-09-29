package handler

import (
	"encoding/xml"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/bjdgyc/anylink/base"
)

// UserState 用于存储用户的登录状态
type UserState struct {
	FailureCount int
	LastAttempt  time.Time
	LockTime     time.Time
}

// 自定义 contextKey 类型，避免键冲突
type contextKey string

// 定义常量作为上下文的键
const loginStatusKey contextKey = "login_status"

// 用户状态映射
var userStates = make(map[string]*UserState)
var mu sync.Mutex

func init() {
	go cleanupUserStates()
}

// 清理过期的登录状态
func cleanupUserStates() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		mu.Lock()
		now := time.Now()
		for username, state := range userStates {
			if now.Sub(state.LastAttempt) > time.Duration(base.Cfg.UserStateExpiration)*time.Second {
				delete(userStates, username)
			}
		}
		mu.Unlock()
	}
}

// 防爆破中间件
func antiBruteForce(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 如果最大验证失败次数为0，则不启用防爆破功能
		if base.Cfg.MaxBanCount == 0 {
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}
		r.Body.Close()

		cr := ClientRequest{}
		err = xml.Unmarshal(body, &cr)
		if err != nil {
			http.Error(w, "Failed to parse XML", http.StatusBadRequest)
			return
		}

		username := cr.Auth.Username

		// 更新用户登录状态
		mu.Lock()
		state, exists := userStates[username]
		if !exists {
			state = &UserState{}
			userStates[username] = state
		}
		// 检查是否已超过锁定时间
		if !state.LockTime.IsZero() {
			if time.Now().After(state.LockTime) {
				// 如果已经超过了锁定时间，重置失败计数和锁定时间
				state.FailureCount = 0
				state.LockTime = time.Time{}
			} else {
				// 如果还在锁定时间内，返回错误信息
				http.Error(w, "Account locked due to too many failed attempts. Try again later.", http.StatusTooManyRequests)
				mu.Unlock()
				return
			}
		}

		// 如果超过时间窗口，重置失败计数
		if time.Since(state.LastAttempt) > time.Duration(base.Cfg.BanResetTime)*time.Second {
			state.FailureCount = 0
		}

		state.LastAttempt = time.Now()
		mu.Unlock()

		// 重新设置请求体以便后续处理器可以访问
		r.Body = io.NopCloser(strings.NewReader(string(body)))

		// 调用下一个处理器
		next.ServeHTTP(w, r)

		// 从 context 中获取登录状态
		loginStatus, ok := r.Context().Value(loginStatusKey).(bool)
		if !ok {
			// 如果没有找到登录状态，默认为登录失败
			loginStatus = false
		}

		// 更新用户登录状态
		mu.Lock()
		defer mu.Unlock()

		if !loginStatus {
			state.FailureCount++
			if state.FailureCount >= base.Cfg.MaxBanCount {
				state.LockTime = time.Now().Add(time.Duration(base.Cfg.LockTime) * time.Second)
			}
		} else {
			state.FailureCount = 0 // 成功登录后重置
		}
	})
}
