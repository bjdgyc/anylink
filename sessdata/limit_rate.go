package sessdata

import (
	"context"
	"fmt"
	"time"

	"github.com/bjdgyc/anylink/common"

	"golang.org/x/time/rate"
)

var Sess = &ConnSession{}

func init() {
	return
	tick := time.Tick(time.Second * 2)
	go func() {
		for range tick {
			uP := common.HumanByte(float64(Sess.BandwidthUpPeriod / BandwidthPeriodSec))
			dP := common.HumanByte(float64(Sess.BandwidthDownPeriod / BandwidthPeriodSec))
			uA := common.HumanByte(float64(Sess.BandwidthUpAll))
			dA := common.HumanByte(float64(Sess.BandwidthDownAll))

			fmt.Printf("rateUp:%s rateDown:%s allUp %s allDown %s \n",
				uP, dP, uA, dA)
		}
	}()
}

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
