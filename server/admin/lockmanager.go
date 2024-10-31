package admin

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/bjdgyc/anylink/base"
)

type LockInfo struct {
	Description string     `json:"description"` // 锁定原因
	Username    string     `json:"username"`    // 用户名
	IP          string     `json:"ip"`          // IP 地址
	State       *LockState `json:"state"`       // 锁定状态信息
}
type LockState struct {
	Locked       bool      `json:"locked"`      // 是否锁定
	FailureCount int       `json:"attempts"`    // 失败次数
	LockTime     time.Time `json:"lock_time"`   // 锁定截止时间
	LastAttempt  time.Time `json:"lastAttempt"` // 最后一次尝试的时间
}
type IPWhitelists struct {
	IP   net.IP
	CIDR *net.IPNet
}

type LockManager struct {
	mu            sync.Mutex
	LoginStatus   sync.Map                         // 登录状态
	ipLocks       map[string]*LockState            // 全局IP锁定状态
	userLocks     map[string]*LockState            // 全局用户锁定状态
	ipUserLocks   map[string]map[string]*LockState // 单用户IP锁定状态
	ipWhitelists  []IPWhitelists                   // 全局IP白名单，包含IP地址和CIDR范围
	cleanupTicker *time.Ticker
}

var lockmanager *LockManager
var once sync.Once

func GetLockManager() *LockManager {
	once.Do(func() {
		lockmanager = &LockManager{
			LoginStatus:  sync.Map{},
			ipLocks:      make(map[string]*LockState),
			userLocks:    make(map[string]*LockState),
			ipUserLocks:  make(map[string]map[string]*LockState),
			ipWhitelists: make([]IPWhitelists, 0),
		}
	})
	return lockmanager
}

const defaultGlobalLockStateExpirationTime = 3600

func InitLockManager() {
	lm := GetLockManager()
	if base.Cfg.AntiBruteForce {
		if base.Cfg.GlobalLockStateExpirationTime <= 0 {
			base.Cfg.GlobalLockStateExpirationTime = defaultGlobalLockStateExpirationTime
		}
		lm.StartCleanupTicker()
		lm.InitIPWhitelist()
	}
}

func GetLocksInfo(w http.ResponseWriter, r *http.Request) {
	lm := GetLockManager()
	locksInfo := lm.GetLocksInfo()

	RespSucess(w, locksInfo)
}

func UnlockUser(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		RespError(w, RespInternalErr, err)
		return
	}
	lockinfo := LockInfo{}
	if err := json.Unmarshal(body, &lockinfo); err != nil {
		RespError(w, RespInternalErr, err)
		return
	}

	if lockinfo.State == nil {
		RespError(w, RespInternalErr, fmt.Errorf("未找到锁定用户！"))
		return
	}
	lm := GetLockManager()

	lm.mu.Lock()
	defer lm.mu.Unlock()

	// 根据用户名和IP查找锁定状态
	var state *LockState
	switch {
	case lockinfo.IP == "" && lockinfo.Username != "":
		state = lm.userLocks[lockinfo.Username] // 全局用户锁定
	case lockinfo.Username != "" && lockinfo.IP != "":
		if userIPMap, exists := lm.ipUserLocks[lockinfo.Username]; exists {
			state = userIPMap[lockinfo.IP] // 单用户 IP 锁定
		}
	default:
		state = lm.ipLocks[lockinfo.IP] // 全局 IP 锁定
	}

	if state == nil || !state.Locked {
		RespError(w, RespInternalErr, fmt.Errorf("锁定状态未找到或已解锁"))
		return
	}

	lm.Unlock(state)
	base.Info("解锁成功:", lockinfo.Description, lockinfo.Username, lockinfo.IP)

	RespSucess(w, "解锁成功！")
}

func (lm *LockManager) GetLocksInfo() []LockInfo {
	var locksInfo []LockInfo

	lm.mu.Lock()
	defer lm.mu.Unlock()

	for ip, state := range lm.ipLocks {
		if state.Locked {
			info := LockInfo{
				Description: "全局IP锁定",
				Username:    "",
				IP:          ip,
				State: &LockState{
					Locked:       state.Locked,
					FailureCount: state.FailureCount,
					LockTime:     state.LockTime,
					LastAttempt:  state.LastAttempt,
				},
			}
			locksInfo = append(locksInfo, info)
		}
	}

	for username, state := range lm.userLocks {
		if state.Locked {
			info := LockInfo{
				Description: "全局用户锁定",
				Username:    username,
				IP:          "",
				State: &LockState{
					Locked:       state.Locked,
					FailureCount: state.FailureCount,
					LockTime:     state.LockTime,
					LastAttempt:  state.LastAttempt,
				},
			}
			locksInfo = append(locksInfo, info)
		}
	}

	for username, ipStates := range lm.ipUserLocks {
		for ip, state := range ipStates {
			if state.Locked {
				info := LockInfo{
					Description: "单用户IP锁定",
					Username:    username,
					IP:          ip,
					State: &LockState{
						Locked:       state.Locked,
						FailureCount: state.FailureCount,
						LockTime:     state.LockTime,
						LastAttempt:  state.LastAttempt,
					},
				}
				locksInfo = append(locksInfo, info)
			}
		}
	}
	return locksInfo
}

// 初始化IP白名单
func (lm *LockManager) InitIPWhitelist() {
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
func (lm *LockManager) IsWhitelisted(ip string) bool {
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

func (lm *LockManager) StartCleanupTicker() {
	lm.cleanupTicker = time.NewTicker(1 * time.Minute)
	go func() {
		for range lm.cleanupTicker.C {
			lm.CleanupExpiredLocks()
		}
	}()
}

// 定期清理过期的锁定
func (lm *LockManager) CleanupExpiredLocks() {
	now := time.Now()
	lm.mu.Lock()
	defer lm.mu.Unlock()

	for ip, state := range lm.ipLocks {
		if !lm.CheckLockState(state, now, base.Cfg.GlobalIPLockTime) ||
			now.Sub(state.LastAttempt) > time.Duration(base.Cfg.GlobalLockStateExpirationTime)*time.Second {
			delete(lm.ipLocks, ip)
		}
	}

	for user, state := range lm.userLocks {
		if !lm.CheckLockState(state, now, base.Cfg.GlobalUserLockTime) ||
			now.Sub(state.LastAttempt) > time.Duration(base.Cfg.GlobalLockStateExpirationTime)*time.Second {
			delete(lm.userLocks, user)
		}
	}

	for user, ipMap := range lm.ipUserLocks {
		for ip, state := range ipMap {
			if !lm.CheckLockState(state, now, base.Cfg.LockTime) ||
				now.Sub(state.LastAttempt) > time.Duration(base.Cfg.GlobalLockStateExpirationTime)*time.Second {
				delete(ipMap, ip)
				if len(ipMap) == 0 {
					delete(lm.ipUserLocks, user)
				}
			}
		}
	}
}

// 检查全局 IP 锁定
func (lm *LockManager) CheckGlobalIPLock(ip string, now time.Time) bool {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	state, exists := lm.ipLocks[ip]
	if !exists {
		return false
	}

	return lm.CheckLockState(state, now, base.Cfg.GlobalIPBanResetTime)
}

// 检查全局用户锁定
func (lm *LockManager) CheckGlobalUserLock(username string, now time.Time) bool {
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
	return lm.CheckLockState(state, now, base.Cfg.GlobalUserBanResetTime)
}

// 检查单个用户的 IP 锁定
func (lm *LockManager) CheckUserIPLock(username, ip string, now time.Time) bool {
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

	return lm.CheckLockState(state, now, base.Cfg.BanResetTime)
}

// 更新全局 IP 锁定状态
func (lm *LockManager) UpdateGlobalIPLock(ip string, now time.Time, success bool) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	state, exists := lm.ipLocks[ip]
	if !exists {
		state = &LockState{}
		lm.ipLocks[ip] = state
	}

	lm.UpdateLockState(state, now, success, base.Cfg.MaxGlobalIPBanCount, base.Cfg.GlobalIPLockTime)
}

// 更新全局用户锁定状态
func (lm *LockManager) UpdateGlobalUserLock(username string, now time.Time, success bool) {
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

	lm.UpdateLockState(state, now, success, base.Cfg.MaxGlobalUserBanCount, base.Cfg.GlobalUserLockTime)
}

// 更新单个用户的 IP 锁定状态
func (lm *LockManager) UpdateUserIPLock(username, ip string, now time.Time, success bool) {
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

	lm.UpdateLockState(state, now, success, base.Cfg.MaxBanCount, base.Cfg.LockTime)
}

// 更新锁定状态
func (lm *LockManager) UpdateLockState(state *LockState, now time.Time, success bool, maxBanCount, lockTime int) {
	if success {
		lm.Unlock(state) // 成功登录后解锁
	} else {
		state.FailureCount++
		if state.FailureCount >= maxBanCount {
			state.LockTime = now.Add(time.Duration(lockTime) * time.Second)
			state.Locked = true // 超过阈值时锁定
		}
	}
	state.LastAttempt = now
}

// 检查锁定状态
func (lm *LockManager) CheckLockState(state *LockState, now time.Time, resetTime int) bool {
	if state == nil || state.LastAttempt.IsZero() {
		return false
	}

	// 如果超过锁定时间，重置锁定状态
	if !state.LockTime.IsZero() && now.After(state.LockTime) {
		lm.Unlock(state) // 锁定期过后解锁
		return false
	}
	// 如果超过窗口时间，重置失败计数
	if now.Sub(state.LastAttempt) > time.Duration(resetTime)*time.Second {
		state.FailureCount = 0
		return false
	}
	return state.Locked
}

// 解锁
func (lm *LockManager) Unlock(state *LockState) {
	state.FailureCount = 0
	state.LockTime = time.Time{}
	state.Locked = false
}
