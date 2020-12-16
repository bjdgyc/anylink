package sessdata

import (
	"context"

	"golang.org/x/time/rate"
)

type LimitRater struct {
	limit *rate.Limiter
}

// lim: 令牌产生速率
// burst: 允许的最大爆发速率
func NewLimitRater(lim, burst int) *LimitRater {
	limit := rate.NewLimiter(rate.Limit(lim), burst)
	return &LimitRater{limit: limit}
}

// bt 不能超过burst大小
func (l *LimitRater) Wait(bt int) error {
	return l.limit.WaitN(context.Background(), bt)
}
