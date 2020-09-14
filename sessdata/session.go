package sessdata

import (
	"crypto/md5"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bjdgyc/anylink/common"
)

const BandwidthPeriodSec = 2 // 流量速率统计周期(秒)

var (
	// session_token -> SessUser
	sessions = sync.Map{} // make(map[string]*Session)
)

// 连接sess
type ConnSession struct {
	Sess                *Session
	MasterSecret        string // dtls协议的 master_secret
	Ip                  net.IP // 分配的ip地址
	LocalIp             net.IP
	MacHw               net.HardwareAddr // 客户端mac地址,从Session取出
	RemoteAddr          string
	Mtu                 int
	TunName             string
	Limit               *LimitRater
	BandwidthUp         uint32 // 使用上行带宽 Byte
	BandwidthDown       uint32 // 使用下行带宽 Byte
	BandwidthUpPeriod   uint32 // 前一周期的总量
	BandwidthDownPeriod uint32
	BandwidthUpAll      uint32 // 使用上行带宽总量
	BandwidthDownAll    uint32 // 使用下行带宽总量
	closeOnce           sync.Once
	CloseChan           chan struct{}
	PayloadIn           chan *Payload
	PayloadOut          chan *Payload
	PayloadArp          chan *Payload
}

type Session struct {
	mux            sync.Mutex
	Sid            string // auth返回的 session-id
	Token          string // session信息的唯一token
	DtlsSid        string // dtls协议的 session_id
	MacAddr        string // 客户端mac地址
	UniqueIdGlobal string // 客户端唯一标示
	UserName       string // 用户名

	LastLogin time.Time
	IsActive  bool

	// 开启link需要设置的参数
	CSess *ConnSession
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func checkSession() {

	// 检测过期的session
	go func() {
		if common.ServerCfg.SessionTimeout == 0 {
			return
		}
		timeout := time.Duration(common.ServerCfg.SessionTimeout) * time.Second
		tick := time.Tick(time.Second * 60)
		for range tick {
			t := time.Now()

			sessions.Range(func(key, value interface{}) bool {
				v := value.(*Session)
				v.mux.Lock()
				defer v.mux.Unlock()

				if v.IsActive == true {
					return true
				}
				if t.Sub(v.LastLogin) > timeout {
					sessions.Delete(key)
				}
				return true
			})

		}
	}()
}

func NewSession() *Session {
	// 生成32位的 token
	btoken := make([]byte, 32)
	rand.Read(btoken)

	// 生成 dtlsn session_id
	dtlsid := make([]byte, 32)
	rand.Read(dtlsid)

	token := fmt.Sprintf("%x", btoken)
	sess := &Session{
		Sid:       fmt.Sprintf("%d", time.Now().Unix()),
		Token:     token,
		DtlsSid:   fmt.Sprintf("%x", dtlsid),
		LastLogin: time.Now(),
	}

	sessions.Store(token, sess)
	return sess
}

func (s *Session) NewConn() *ConnSession {
	s.mux.Lock()
	active := s.IsActive
	macAddr := s.MacAddr
	s.mux.Unlock()
	if active == true {
		s.CSess.Close()
	}

	limit := LimitClient(s.UserName, false)
	if limit == false {
		return nil
	}
	// 获取客户端mac地址
	macHw, err := net.ParseMAC(macAddr)
	if err != nil {
		sum := md5.Sum([]byte(s.UniqueIdGlobal))
		macHw = sum[8:13] // 5个byte
		macHw = append([]byte{0x00}, macHw...)
		macAddr = macHw.String()
	}
	ip := AcquireIp(macAddr)
	if ip == nil {
		return nil
	}

	cSess := &ConnSession{
		Sess:       s,
		MacHw:      macHw,
		Ip:         ip,
		closeOnce:  sync.Once{},
		CloseChan:  make(chan struct{}),
		PayloadIn:  make(chan *Payload),
		PayloadOut: make(chan *Payload),
		PayloadArp: make(chan *Payload, 1000),
		// Limit:      NewLimitRater(1024 * 1024),
	}

	go cSess.ratePeriod()

	s.mux.Lock()
	s.MacAddr = macAddr
	s.IsActive = true
	s.CSess = cSess
	s.mux.Unlock()
	return cSess
}

func (cs *ConnSession) Close() {
	cs.closeOnce.Do(func() {
		log.Println("closeOnce:", cs.Ip)
		cs.Sess.mux.Lock()
		defer cs.Sess.mux.Unlock()

		close(cs.CloseChan)
		cs.Sess.IsActive = false
		cs.Sess.LastLogin = time.Now()
		cs.Sess.CSess = nil

		ReleaseIp(cs.Ip, cs.Sess.MacAddr)
		LimitClient(cs.Sess.UserName, true)
	})
}

func (cs *ConnSession) ratePeriod() {
	tick := time.Tick(time.Second * BandwidthPeriodSec)
	for range tick {
		select {
		case <-cs.CloseChan:
			return
		default:
		}

		// 实时流量清零
		rtUp := atomic.SwapUint32(&cs.BandwidthUp, 0)
		rtDown := atomic.SwapUint32(&cs.BandwidthDown, 0)
		// 设置上一周期的流量
		atomic.SwapUint32(&cs.BandwidthUpPeriod, rtUp)
		atomic.SwapUint32(&cs.BandwidthDownPeriod, rtDown)
		// 累加所有流量
		atomic.AddUint32(&cs.BandwidthUpAll, rtUp)
		atomic.AddUint32(&cs.BandwidthDownAll, rtDown)
	}
}

const MaxMtu = 1460

func (cs *ConnSession) SetMtu(mtu string) {
	cs.Mtu = MaxMtu

	mi, err := strconv.Atoi(mtu)
	if err != nil || mi < 100 {
		return
	}

	if mi < MaxMtu {
		cs.Mtu = mi
	}
}

func (cs *ConnSession) RateLimit(byt int, isUp bool) error {
	if isUp {
		atomic.AddUint32(&cs.BandwidthUp, uint32(byt))
		return nil
	}
	// 只对下行速率限制
	atomic.AddUint32(&cs.BandwidthDown, uint32(byt))
	if cs.Limit == nil {
		return nil
	}
	return cs.Limit.Wait(byt)
}

func SToken2Sess(stoken string) *Session {
	stoken = strings.TrimSpace(stoken)
	sarr := strings.Split(stoken, "@")
	token := sarr[1]

	if sess, ok := sessions.Load(token); ok {
		return sess.(*Session)
	}

	return nil
}

func Dtls2Sess(dtlsid []byte) *Session {
	return nil
}

func DelSess(token string) {
	// sessions.Delete(token)
}

func DelSessByStoken(stoken string) {
	stoken = strings.TrimSpace(stoken)
	sarr := strings.Split(stoken, "@")
	token := sarr[1]
	sessions.Delete(token)
}
