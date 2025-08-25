package admin

import (
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/bjdgyc/anylink/base"
	"github.com/stretchr/testify/assert"
)

// Helper function to reset the singleton for test isolation
func resetLockManager() {
	once = sync.Once{}
	lockmanager = nil
}

// 测试 GetLockManager 函数
func TestGetLockManager(t *testing.T) {
	resetLockManager()
	base.Test()
	setupTestConfig()

	t.Run("Singleton_Pattern", func(t *testing.T) {
		lm1 := GetLockManager()
		lm2 := GetLockManager()
		assert.Same(t, lm1, lm2, "GetLockManager应该返回同一个实例")
	})
}

// 测试并发竞争条件
func TestLockManager_RaceConditions(t *testing.T) {
	resetLockManager()
	base.Test()
	setupTestConfig()

	lm := GetLockManager()

	t.Run("Concurrent_CheckAndUpdate", func(t *testing.T) {
		username := "raceuser"
		ipaddr := "192.168.1.10:12345"

		var wg sync.WaitGroup
		results := make([]bool, 20)

		// 并发执行检查和更新操作
		for i := 0; i < 20; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				// 交替执行检查和更新操作
				if index%2 == 0 {
					results[index] = lm.CheckLocked(username, ipaddr)
				} else {
					lm.UpdateLoginStatus(username, ipaddr, false)
				}
			}(i)
		}

		wg.Wait()

		// 验证最终状态的一致性
		finalResult := lm.CheckLocked(username, ipaddr)
		assert.False(t, finalResult, "高并发后应该被锁定")
	})

	t.Run("Concurrent_MultipleUsers", func(t *testing.T) {
		ipaddr := "192.168.1.11:12345"
		var wg sync.WaitGroup

		// 多个用户同时从同一IP进行攻击
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func(userIndex int) {
				defer wg.Done()
				username := fmt.Sprintf("user%d", userIndex)
				lm.UpdateLoginStatus(username, ipaddr, false)
			}(i)
		}

		wg.Wait()

		// 验证全局IP锁定是否正确触发
		result := lm.CheckLocked("newuser", ipaddr)
		assert.False(t, result, "多用户并发攻击后IP应该被全局锁定")
	})

	t.Run("Concurrent_CleanupAndUpdate", func(t *testing.T) {
		username := "cleanuprace"
		ipaddr := "192.168.1.12:12345"

		// 先创建一些状态
		lm.UpdateLoginStatus(username, ipaddr, false)

		var wg sync.WaitGroup

		// 并发执行清理和更新操作
		wg.Add(2)
		go func() {
			defer wg.Done()
			lm.CleanupExpiredLocks()
		}()

		go func() {
			defer wg.Done()
			lm.UpdateLoginStatus(username, ipaddr, false)
		}()

		wg.Wait()

		// 验证状态一致性
		result := lm.CheckLocked(username, ipaddr)
		// 结果可能是true或false，但不应该panic
		_ = result
	})
}

// 测试 CheckLocked 函数
func TestCheckLocked(t *testing.T) {
	resetLockManager()
	base.Test()
	setupTestConfig()

	lm := GetLockManager()

	t.Run("InitialState_AllowLogin", func(t *testing.T) {
		result := lm.CheckLocked("testuser", "192.168.1.1:12345")
		assert.True(t, result, "初始状态应该允许登录")
	})

	t.Run("LockedState_DenyLogin", func(t *testing.T) {
		username := "testuser"
		ipaddr := "192.168.1.1:12345"

		// 模拟5次失败登录
		for i := 0; i < 5; i++ {
			lm.UpdateLoginStatus(username, ipaddr, false)
		}

		result := lm.CheckLocked(username, ipaddr)
		assert.False(t, result, "5次失败后应该被锁定")
	})
}

// 测试 UpdateLoginStatus 函数
func TestUpdateLoginStatus(t *testing.T) {
	resetLockManager()
	base.Test()
	setupTestConfig()

	lm := GetLockManager()

	t.Run("FailureCount_Increment", func(t *testing.T) {
		username := "testuser"
		ipaddr := "192.168.1.1:12345"

		// 测试失败计数递增
		for i := 1; i <= 3; i++ {
			lm.UpdateLoginStatus(username, ipaddr, false)

			lm.mu.Lock()
			userIPMap := lm.ipUserLocks[username]
			ip, _, _ := net.SplitHostPort(ipaddr)
			state := userIPMap[ip]
			assert.Equal(t, i, state.FailureCount, fmt.Sprintf("第%d次失败后计数应该为%d", i, i))
			lm.mu.Unlock()
		}
	})

	t.Run("SuccessLogin_ResetCount", func(t *testing.T) {
		username := "successuser"
		ipaddr := "192.168.1.2:12345"

		// 先失败几次
		for i := 0; i < 3; i++ {
			lm.UpdateLoginStatus(username, ipaddr, false)
		}

		// 成功登录
		lm.UpdateLoginStatus(username, ipaddr, true)

		// 验证计数被重置
		lm.mu.Lock()
		userIPMap := lm.ipUserLocks[username]
		ip, _, _ := net.SplitHostPort(ipaddr)
		state := userIPMap[ip]
		assert.Equal(t, 0, state.FailureCount, "成功登录应该重置失败计数")
		lm.mu.Unlock()
	})
}

// 测试 UpdateLockState 函数
func TestUpdateLockState(t *testing.T) {
	resetLockManager()
	base.Test()
	setupTestConfig()

	lm := GetLockManager()

	t.Run("LockedState_NoUpdate", func(t *testing.T) {
		// 测试已锁定状态不会被重复更新
		username := "lockeduser"
		ipaddr := "192.168.1.2:12345"

		// 先锁定用户
		for i := 0; i < 5; i++ {
			lm.UpdateLoginStatus(username, ipaddr, false)
		}

		// 获取当前状态
		lm.mu.Lock()
		userIPMap := lm.ipUserLocks[username]
		ip, _, _ := net.SplitHostPort(ipaddr)
		state := userIPMap[ip]
		originalCount := state.FailureCount
		originalLastAttempt := state.LastAttempt
		lm.mu.Unlock()

		// 尝试再次更新
		lm.UpdateLoginStatus(username, ipaddr, false)

		// 验证状态没有改变
		lm.mu.Lock()
		newState := lm.ipUserLocks[username][ip]
		assert.Equal(t, originalCount, newState.FailureCount, "已锁定状态的失败计数不应该改变")
		assert.Equal(t, originalLastAttempt, newState.LastAttempt, "已锁定状态的时间戳不应该改变")
		lm.mu.Unlock()
	})
}

// 测试 CheckGlobalIPLock 函数
func TestCheckGlobalIPLock(t *testing.T) {
	resetLockManager()
	base.Test()
	setupTestConfig()

	lm := GetLockManager()

	t.Run("GlobalIP_Protection", func(t *testing.T) {
		ipaddr := "192.168.1.3:12345"
		ip, _, _ := net.SplitHostPort(ipaddr)

		// 使用不同用户名从同一IP进行攻击
		for i := 0; i < 40; i++ {
			username := fmt.Sprintf("user%d", i)
			lm.UpdateLoginStatus(username, ipaddr, false)
		}

		// 验证全局IP锁定检查
		result := lm.CheckGlobalIPLock(ip, time.Now())
		assert.True(t, result, "IP应该被全局锁定")
	})
}

// 测试 CheckGlobalUserLock 函数
func TestCheckGlobalUserLock(t *testing.T) {
	resetLockManager()
	base.Test()
	setupTestConfig()

	lm := GetLockManager()

	t.Run("GlobalUser_Protection", func(t *testing.T) {
		username := "globaluser"

		// 同一用户从不同IP进行攻击
		for i := 0; i < 20; i++ {
			ipaddr := fmt.Sprintf("192.168.1.%d:12345", 100+i)
			lm.UpdateLoginStatus(username, ipaddr, false)
		}

		// 验证全局用户锁定检查
		result := lm.CheckGlobalUserLock(username, time.Now())
		assert.True(t, result, "用户应该被全局锁定")
	})
}

// 测试 CheckUserIPLock 函数
func TestCheckUserIPLock(t *testing.T) {
	resetLockManager()
	base.Test()
	setupTestConfig()

	lm := GetLockManager()

	t.Run("UserIP_Protection", func(t *testing.T) {
		username := "useripuser"
		ipaddr := "192.168.1.4:12345"
		ip, _, _ := net.SplitHostPort(ipaddr)

		// 单用户IP锁定测试
		for i := 0; i < 5; i++ {
			lm.UpdateLoginStatus(username, ipaddr, false)
		}

		// 验证单用户IP锁定检查
		result := lm.CheckUserIPLock(username, ip, time.Now())
		assert.True(t, result, "单用户IP应该被锁定")
	})
}

// 测试 InitIPList 和 IsInIPList 函数
func TestInitIPList_IsInIPList(t *testing.T) {
	resetLockManager()
	base.Test()
	setupTestConfig()

	lm := GetLockManager()

	t.Run("Whitelist_Functionality", func(t *testing.T) {
		// 手动初始化IP列表
		lm.InitIPList(IPListWhite, base.Cfg.IPWhiteList)

		// 测试白名单检查
		result := lm.IsInIPList("192.168.90.1", IPListWhite)
		assert.True(t, result, "192.168.90.1应该在白名单中")

		// 测试CIDR范围
		result2 := lm.IsInIPList("172.16.0.100", IPListWhite)
		assert.True(t, result2, "172.16.0.100应该在CIDR范围内")
	})

	t.Run("Blacklist_Functionality", func(t *testing.T) {
		// 手动初始化黑名单
		lm.InitIPList(IPListBlack, base.Cfg.IPBlackList)

		// 测试黑名单检查
		result := lm.IsInIPList("10.0.0.1", IPListBlack)
		assert.True(t, result, "10.0.0.1应该在黑名单中")
	})
}

// 测试 GetLocksInfo 函数
func TestGetLocksInfo(t *testing.T) {
	resetLockManager()
	base.Test()
	setupTestConfig()

	lm := GetLockManager()

	t.Run("EmptyState", func(t *testing.T) {
		locksInfo := lm.GetLocksInfo()
		assert.Empty(t, locksInfo, "初始状态应该没有锁定信息")
	})

	t.Run("WithLocks", func(t *testing.T) {
		// 创建锁定状态
		username := "testuser"
		ipaddr := "192.168.1.5:12345"

		for i := 0; i < 5; i++ {
			lm.UpdateLoginStatus(username, ipaddr, false)
		}

		locksInfo := lm.GetLocksInfo()
		assert.NotEmpty(t, locksInfo, "应该有锁定信息")
	})
}

// 测试 CleanupExpiredLocks 函数
func TestCleanupExpiredLocks(t *testing.T) {
	resetLockManager()
	base.Test()
	setupTestConfig()

	lm := GetLockManager()

	t.Run("ExpiredLocks_Cleanup", func(t *testing.T) {
		username := "cleanupuser"
		ipaddr := "192.168.1.6:12345"

		// 创建锁定状态
		lm.UpdateLoginStatus(username, ipaddr, false)

		// 模拟过期状态
		lm.mu.Lock()
		userIPMap := lm.ipUserLocks[username]
		ip, _, _ := net.SplitHostPort(ipaddr)
		state := userIPMap[ip]
		state.LastAttempt = time.Now().Add(-7200 * time.Second) // 2小时前
		lm.mu.Unlock()

		// 执行清理
		lm.CleanupExpiredLocks()

		// 验证过期状态被清理
		lm.mu.Lock()
		_, exists := lm.ipUserLocks[username]
		lm.mu.Unlock()
		assert.False(t, exists, "过期的锁定状态应该被清理")
	})
}

// 测试 CheckLockState 函数
func TestCheckLockState(t *testing.T) {
	resetLockManager()
	base.Test()
	setupTestConfig()

	lm := GetLockManager()

	t.Run("TimeWindow_Reset", func(t *testing.T) {
		state := &LockState{
			FailureCount: 3,
			LastAttempt:  time.Now().Add(-700 * time.Second), // 700秒前
		}

		// 检查状态（应该重置计数）
		result := lm.CheckLockState(state, time.Now(), 600) // 600秒重置时间

		assert.False(t, result, "超过重置时间应该返回false")
		assert.Equal(t, 0, state.FailureCount, "失败计数应该被重置")
	})
}

// 辅助函数
func setupTestConfig() {
	base.Cfg.AntiBruteForce = true
	base.Cfg.IPWhiteList = "192.168.90.1,172.16.0.0/24"
	base.Cfg.IPBlackList = "10.0.0.1"
	base.Cfg.MaxBanCount = 5
	base.Cfg.BanResetTime = 600
	base.Cfg.LockTime = 300
	base.Cfg.MaxGlobalUserBanCount = 20
	base.Cfg.GlobalUserBanResetTime = 600
	base.Cfg.GlobalUserLockTime = 300
	base.Cfg.MaxGlobalIPBanCount = 40
	base.Cfg.GlobalIPBanResetTime = 1200
	base.Cfg.GlobalIPLockTime = 300
	base.Cfg.GlobalLockStateExpirationTime = 3600
}
