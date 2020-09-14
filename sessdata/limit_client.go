package sessdata

import (
	"sync"

	"github.com/bjdgyc/anylink/common"
)

const limitAllKey = "__ALL__"

var (
	limitClient = map[string]int{limitAllKey: 0}
	limitMux    = sync.Mutex{}
)

func LimitClient(user string, close bool) bool {
	limitMux.Lock()
	defer limitMux.Unlock()
	// defer fmt.Println(limitClient)

	_all := limitClient[limitAllKey]
	c, ok := limitClient[user]
	if !ok { // 不存在用户
		limitClient[user] = 0
	}

	if close {
		limitClient[user] = c - 1
		limitClient[limitAllKey] = _all - 1
		return true
	}

	// 全局判断
	if _all >= common.ServerCfg.MaxClient {
		return false
	}

	// 超出同一个用户限制
	if c >= common.ServerCfg.MaxUserClient {
		return false
	}

	limitClient[user] = c + 1
	limitClient[limitAllKey] = _all + 1
	return true
}
