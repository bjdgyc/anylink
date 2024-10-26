package handler

import (
	"encoding/xml"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/bjdgyc/anylink/base"
)

const loginStatusKey = "login_status"
const defaultGlobalLockStateExpirationTime = 3600

func initAntiBruteForce() {
	if base.Cfg.AntiBruteForce {
		if base.Cfg.GlobalLockStateExpirationTime <= 0 {
			base.Cfg.GlobalLockStateExpirationTime = defaultGlobalLockStateExpirationTime
		}
		lockManager.startCleanupTicker()
		lockManager.initIPWhitelist()
	}
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
		if lockManager.isWhitelisted(ip) {
			r.Body = io.NopCloser(strings.NewReader(string(body)))
			next.ServeHTTP(w, r)
			return
		}

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
			base.Warn("IP", ip, "is globally locked. Try again later.")
			http.Error(w, "Account globally locked due to too many failed attempts. Try again later.", http.StatusTooManyRequests)
			return
		}

		// 检查全局用户锁定
		if base.Cfg.MaxGlobalUserBanCount > 0 && lockManager.checkGlobalUserLock(username, now) {
			base.Warn("User", username, "is globally locked. Try again later.")
			http.Error(w, "Account globally locked due to too many failed attempts. Try again later.", http.StatusTooManyRequests)
			return
		}

		// 检查单个用户的 IP 锁定
		if base.Cfg.MaxBanCount > 0 && lockManager.checkUserIPLock(username, ip, now) {
			base.Warn("IP", ip, "is locked for user", username, "Try again later.")
			http.Error(w, "Account locked due to too many failed attempts. Try again later.", http.StatusTooManyRequests)
			return
		}

		// 重新设置请求体以便后续处理器可以访问
		r.Body = io.NopCloser(strings.NewReader(string(body)))

		// 调用下一个处理器
		next.ServeHTTP(w, r)

		// 检查登录状态
		Status, _ := lockManager.loginStatus.Load(loginStatusKey)
		loginStatus, _ := Status.(bool)

		// 更新用户登录状态
		lockManager.updateGlobalIPLock(ip, now, loginStatus)
		lockManager.updateGlobalUserLock(username, now, loginStatus)
		lockManager.updateUserIPLock(username, ip, now, loginStatus)

		// 清除登录状态
		lockManager.loginStatus.Delete(loginStatusKey)
	})
}

type LockState struct {
	FailureCount int
	LockTime     time.Time
	LastAttempt  time.Time
}
type IPWhitelists struct {
	IP   net.IP
	CIDR *net.IPNet
}

type LockManager struct {
	mu           sync.Mutex
	loginStatus  sync.Map                         // 登录状态
	ipLocks      map[string]*LockState            // 全局IP锁定状态
	userLocks    map[string]*LockState            // 全局用户锁定状态
	ipUserLocks  map[string]map[string]*LockState // 单用户IP锁定状态
	ipWhitelists []IPWhitelists                   // 全局IP白名单，包含IP地址和CIDR范围
	// rateLimiter   map[string]*rate.Limiter         // 速率限制器
	cleanupTicker *time.Ticker
}

var lockManager = &LockManager{
	loginStatus:  sync.Map{},
	ipLocks:      make(map[string]*LockState),
	userLocks:    make(map[string]*LockState),
	ipUserLocks:  make(map[string]map[string]*LockState),
	ipWhitelists: make([]IPWhitelists, 0),
	// rateLimiter: make(map[string]*rate.Limiter),
}

// 初始化IP白名单
func (lm *LockManager) initIPWhitelist() {
	ipWhitelist := strings.Split(base.Cfg.IPWhitelist, ",")
	for _, ipWhitelist := range ipWhitelist {
		ipWhitelist = strings.TrimSpace(ipWhitelist)
		if ipWhitelist == "" {
			continue
		}

		_, ipNet, err := net.ParseCIDR(ipWhitelist)
		if err == nil {
			lm.ipWhitelists = append(lm.ipWhitelists, IPWhitelists{CIDR: ipNet})
			continue
		}

		ip := net.ParseIP(ipWhitelist)
		if ip != nil {
			lm.ipWhitelists = append(lm.ipWhitelists, IPWhitelists{IP: ip})
			continue
		}
	}
}

// 检查 IP 是否在白名单中
func (lm *LockManager) isWhitelisted(ip string) bool {
	clientIP := net.ParseIP(ip)
	if clientIP == nil {
		return false
	}
	for _, ipWhitelist := range lm.ipWhitelists {
		if ipWhitelist.CIDR != nil && ipWhitelist.CIDR.Contains(clientIP) {
			return true
		}
		if ipWhitelist.IP != nil && ipWhitelist.IP.Equal(clientIP) {
			return true
		}
	}
	return false
}

func (lm *LockManager) startCleanupTicker() {
	lm.cleanupTicker = time.NewTicker(5 * time.Minute)
	go func() {
		for range lm.cleanupTicker.C {
			lm.cleanupExpiredLocks()
		}
	}()
}

// 定期清理过期的锁定
func (lm *LockManager) cleanupExpiredLocks() {
	now := time.Now()

	var ipKeys, userKeys []string
	var IPuserKeys []struct{ user, ip string }

	lm.mu.Lock()
	for ip, state := range lm.ipLocks {
		if now.Sub(state.LastAttempt) > time.Duration(base.Cfg.GlobalLockStateExpirationTime)*time.Second {
			ipKeys = append(ipKeys, ip)
		}
	}

	for user, state := range lm.userLocks {
		if now.Sub(state.LastAttempt) > time.Duration(base.Cfg.GlobalLockStateExpirationTime)*time.Second {
			userKeys = append(userKeys, user)
		}
	}

	for user, ipMap := range lm.ipUserLocks {
		for ip, state := range ipMap {
			if now.Sub(state.LastAttempt) > time.Duration(base.Cfg.GlobalLockStateExpirationTime)*time.Second {
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

// 检查全局 IP 锁定
func (lm *LockManager) checkGlobalIPLock(ip string, now time.Time) bool {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	state, exists := lm.ipLocks[ip]
	if !exists {
		return false
	}

	return lm.checkLockState(state, now, base.Cfg.GlobalIPBanResetTime)
}

// 检查全局用户锁定
func (lm *LockManager) checkGlobalUserLock(username string, now time.Time) bool {
	// 我也不知道为什么cisco anyconnect每次连接会先传一个空用户请求····
	if username == "" {
		return false
	}
	lm.mu.Lock()
	defer lm.mu.Unlock()

	state, exists := lm.userLocks[username]
	if !exists {
		return false
	}
	return lm.checkLockState(state, now, base.Cfg.GlobalUserBanResetTime)
}

// 检查单个用户的 IP 锁定
func (lm *LockManager) checkUserIPLock(username, ip string, now time.Time) bool {
	// 我也不知道为什么cisco anyconnect每次连接会先传一个空用户请求····
	if username == "" {
		return false
	}
	lm.mu.Lock()
	defer lm.mu.Unlock()

	userIPMap, userExists := lm.ipUserLocks[username]
	if !userExists {
		return false
	}

	state, ipExists := userIPMap[ip]
	if !ipExists {
		return false
	}

	return lm.checkLockState(state, now, base.Cfg.BanResetTime)
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

	lm.updateLockState(state, now, success, base.Cfg.MaxGlobalIPBanCount, base.Cfg.GlobalIPLockTime)
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

	lm.updateLockState(state, now, success, base.Cfg.MaxGlobalUserBanCount, base.Cfg.GlobalUserLockTime)
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

	lm.updateLockState(state, now, success, base.Cfg.MaxBanCount, base.Cfg.LockTime)
}

// 更新锁定状态
func (lm *LockManager) updateLockState(state *LockState, now time.Time, success bool, maxBanCount, lockTime int) {
	if success {
		state.FailureCount = 0
		state.LockTime = time.Time{}
	} else {
		state.FailureCount++
		if state.FailureCount >= maxBanCount {
			state.LockTime = now.Add(time.Duration(lockTime) * time.Second)
		}
	}
	state.LastAttempt = now
}

// 检查锁定状态
func (lm *LockManager) checkLockState(state *LockState, now time.Time, resetTime int) bool {
	if state == nil || state.LastAttempt.IsZero() {
		return false
	}

	// 如果超过锁定时间，重置锁定状态
	if !state.LockTime.IsZero() && now.After(state.LockTime) {
		state.FailureCount = 0
		state.LockTime = time.Time{}
		return false
	}
	// 如果超过窗口时间，重置失败计数
	if now.Sub(state.LastAttempt) > time.Duration(resetTime)*time.Second {
		state.FailureCount = 0
		state.LockTime = time.Time{}
		return false
	}
	// 如果锁定时间还在有效期内，继续锁定
	if !state.LockTime.IsZero() && now.Before(state.LockTime) {
		return true
	}
	return false
}
