package handler

import (
	"encoding/xml"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/bjdgyc/anylink/base"
)

// 自定义 contextKey 类型，避免键冲突
type contextKey string

// 定义常量作为上下文的键
const loginStatusKey contextKey = "login_status"

func init() {
	lockManager.startCleanupTicker()
}

// 防爆破中间件
func antiBruteForce(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 防爆破功能全局开关
		if !base.Cfg.AntiBruteForce {
			next.ServeHTTP(w, r)
			return
		}

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
		ip, _, err := net.SplitHostPort(r.RemoteAddr) // 提取纯 IP 地址，去掉端口号
		if err != nil {
			http.Error(w, "Unable to parse IP address", http.StatusInternalServerError)
			return
		}

		now := time.Now()

		// // 速率限制
		// lockManager.mu.RLock()
		// limiter, exists := lockManager.rateLimiter[ip]
		// if !exists {
		// 	limiter = rate.NewLimiter(rate.Limit(base.Cfg.RateLimit), base.Cfg.Burst)
		// 	lockManager.rateLimiter[ip] = limiter
		// }
		// lockManager.mu.RUnlock()

		// if !limiter.Allow() {
		// 	log.Printf("Rate limit exceeded for IP %s. Try again later.", ip)
		// 	http.Error(w, "Rate limit exceeded. Try again later.", http.StatusTooManyRequests)
		// 	return
		// }

		// 检查全局 IP 锁定
		if base.Cfg.MaxGlobalIPBanCount > 0 && lockManager.checkGlobalIPLock(ip, now) {
			log.Printf("IP %s is globally locked. Try again later.", ip)
			http.Error(w, "Account globally locked due to too many failed attempts. Try again later.", http.StatusTooManyRequests)
			return
		}

		// 检查全局用户锁定
		if base.Cfg.MaxGlobalUserBanCount > 0 && lockManager.checkGlobalUserLock(username, now) {
			log.Printf("User %s is globally locked. Try again later.", username)
			http.Error(w, "Account globally locked due to too many failed attempts. Try again later.", http.StatusTooManyRequests)
			return
		}

		// 检查单个用户的 IP 锁定
		if base.Cfg.MaxBanCount > 0 && lockManager.checkUserIPLock(username, ip, now) {
			log.Printf("IP %s is locked for user %s. Try again later.", ip, username)
			http.Error(w, "Account locked due to too many failed attempts. Try again later.", http.StatusTooManyRequests)
			return
		}

		// 重新设置请求体以便后续处理器可以访问
		r.Body = io.NopCloser(strings.NewReader(string(body)))

		// 调用下一个处理器
		next.ServeHTTP(w, r)

		// 从 context 中获取登录状态
		loginStatus, _ := r.Context().Value(loginStatusKey).(bool)

		// 更新用户登录状态
		lockManager.updateGlobalIPLock(ip, now, loginStatus)
		lockManager.updateGlobalUserLock(username, now, loginStatus)
		lockManager.updateUserIPLock(username, ip, now, loginStatus)
	})
}

type LockState struct {
	FailureCount int
	LockTime     time.Time
	LastAttempt  time.Time
}

type LockManager struct {
	mu          sync.RWMutex
	ipLocks     map[string]*LockState            // 全局IP锁定状态
	userLocks   map[string]*LockState            // 全局用户锁定状态
	ipUserLocks map[string]map[string]*LockState // 单用户IP锁定状态
	// rateLimiter   map[string]*rate.Limiter         // 速率限制器
	cleanupTicker *time.Ticker
}

var lockManager = &LockManager{
	ipLocks:     make(map[string]*LockState),
	userLocks:   make(map[string]*LockState),
	ipUserLocks: make(map[string]map[string]*LockState),
	// rateLimiter: make(map[string]*rate.Limiter),
}

func (lm *LockManager) startCleanupTicker() {
	lm.cleanupTicker = time.NewTicker(1 * time.Minute)
	go func() {
		for range lm.cleanupTicker.C {
			lm.cleanupExpiredLocks()
		}
	}()
}

// 定期清理过期的锁定
func (lm *LockManager) cleanupExpiredLocks() {
	go func() {
		for range time.Tick(5 * time.Minute) {
			now := time.Now()

			var ipKeys, userKeys []string
			var IPuserKeys []struct{ user, ip string }

			lm.mu.Lock()
			for ip, state := range lm.ipLocks {
				if now.Sub(state.LastAttempt) > time.Duration(base.Cfg.GlobalIPBanResetTime)*time.Second {
					ipKeys = append(ipKeys, ip)
				}
			}

			for user, state := range lm.userLocks {
				if now.Sub(state.LastAttempt) > time.Duration(base.Cfg.GlobalUserBanResetTime)*time.Second {
					userKeys = append(userKeys, user)
				}
			}

			for user, ipMap := range lm.ipUserLocks {
				for ip, state := range ipMap {
					if now.Sub(state.LastAttempt) > time.Duration(base.Cfg.BanResetTime)*time.Second {
						IPuserKeys = append(IPuserKeys, struct{ user, ip string }{user, ip})
					}
				}
			}
			lm.mu.Unlock()

			lm.mu.Lock()
			for _, ip := range ipKeys {
				delete(lm.ipLocks, ip)
			}
			for _, user := range userKeys {
				delete(lm.userLocks, user)
			}
			for _, key := range IPuserKeys {
				delete(lm.ipUserLocks[key.user], key.ip)
				if len(lm.ipUserLocks[key.user]) == 0 {
					delete(lm.ipUserLocks, key.user)
				}
			}
			lm.mu.Unlock()
		}
	}()
}

// 检查全局 IP 锁定
func (lm *LockManager) checkGlobalIPLock(ip string, now time.Time) bool {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	state, exists := lm.ipLocks[ip]
	if !exists {
		return false
	}

	if !state.LockTime.IsZero() && now.Before(state.LockTime) {
		return true
	}

	// 如果超过时间窗口，重置失败计数
	if now.Sub(state.LastAttempt) > time.Duration(base.Cfg.GlobalIPBanResetTime)*time.Second {
		state.FailureCount = 0
		state.LockTime = time.Time{}
	}

	return false
}

// 检查全局用户锁定
func (lm *LockManager) checkGlobalUserLock(username string, now time.Time) bool {
	// 我也不知道为什么cisco anyconnect每次连接会先传一个空用户请求····
	if username == "" {
		return false
	}
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	state, exists := lm.userLocks[username]
	if !exists {
		return false
	}

	if !state.LockTime.IsZero() && now.Before(state.LockTime) {
		return true
	}

	// 如果超过时间窗口，重置失败计数
	if now.Sub(state.LastAttempt) > time.Duration(base.Cfg.GlobalUserBanResetTime)*time.Second {
		state.FailureCount = 0
		state.LockTime = time.Time{}
	}

	return false
}

// 检查单个用户的 IP 锁定
func (lm *LockManager) checkUserIPLock(username, ip string, now time.Time) bool {
	// 我也不知道为什么cisco anyconnect每次连接会先传一个空用户请求····
	if username == "" {
		return false
	}
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	userIPMap, userExists := lm.ipUserLocks[username]
	if !userExists {
		return false
	}

	state, ipExists := userIPMap[ip]
	if !ipExists {
		return false
	}

	if !state.LockTime.IsZero() && now.Before(state.LockTime) {
		return true
	}

	// 如果超过时间窗口，重置失败计数
	if now.Sub(state.LastAttempt) > time.Duration(base.Cfg.BanResetTime)*time.Second {
		state.FailureCount = 0
		state.LockTime = time.Time{}
	}

	return false
}

// 更新全局 IP 锁定状态
func (lm *LockManager) updateGlobalIPLock(ip string, now time.Time, success bool) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	state, exists := lm.ipLocks[ip]
	if !exists {
		state = &LockState{}
		lm.ipLocks[ip] = state
	}

	if success {
		state.FailureCount = 0
		state.LockTime = time.Time{}
	} else {
		state.FailureCount++
		if state.FailureCount >= base.Cfg.MaxGlobalIPBanCount {
			state.LockTime = now.Add(time.Duration(base.Cfg.GlobalIPLockTime) * time.Second)
		}
	}
	state.LastAttempt = now
}

// 更新全局用户锁定状态
func (lm *LockManager) updateGlobalUserLock(username string, now time.Time, success bool) {
	// 我也不知道为什么cisco anyconnect每次连接会先传一个空用户请求····
	if username == "" {
		return
	}
	lm.mu.Lock()
	defer lm.mu.Unlock()

	state, exists := lm.userLocks[username]
	if !exists {
		state = &LockState{}
		lm.userLocks[username] = state
	}

	if success {
		state.FailureCount = 0
		state.LockTime = time.Time{}
	} else {
		state.FailureCount++
		if state.FailureCount >= base.Cfg.MaxGlobalUserBanCount {
			state.LockTime = now.Add(time.Duration(base.Cfg.GlobalUserLockTime) * time.Second)
		}
	}
	state.LastAttempt = now
}

// 更新单个用户的 IP 锁定状态
func (lm *LockManager) updateUserIPLock(username, ip string, now time.Time, success bool) {
	// 我也不知道为什么cisco anyconnect每次连接会先传一个空用户请求····
	if username == "" {
		return
	}
	lm.mu.Lock()
	defer lm.mu.Unlock()

	userIPMap, userExists := lm.ipUserLocks[username]
	if !userExists {
		userIPMap = make(map[string]*LockState)
		lm.ipUserLocks[username] = userIPMap
	}

	state, ipExists := userIPMap[ip]
	if !ipExists {
		state = &LockState{}
		userIPMap[ip] = state
	}

	if success {
		state.FailureCount = 0
		state.LockTime = time.Time{}
	} else {
		state.FailureCount++
		if state.FailureCount >= base.Cfg.MaxBanCount {
			state.LockTime = now.Add(time.Duration(base.Cfg.LockTime) * time.Second)
		}
	}
	state.LastAttempt = now
}
