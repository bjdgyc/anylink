package sessdata

import (
	"crypto/md5"
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bjdgyc/anylink/base"
	"github.com/bjdgyc/anylink/dbdata"
)

var (
	// session_token -> SessUser
	sessions = make(map[string]*Session)
	sessMux  sync.Mutex
)

// 连接sess
type ConnSession struct {
	Sess                *Session
	MasterSecret        string // dtls协议的 master_secret
	IpAddr              net.IP // 分配的ip地址
	LocalIp             net.IP
	MacHw               net.HardwareAddr // 客户端mac地址,从Session取出
	RemoteAddr          string
	Mtu                 int
	TunName             string
	Client              string // 客户端  mobile pc
	CstpDpd             int
	Group               *dbdata.Group
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
}

type Session struct {
	mux            sync.Mutex
	Sid            string // auth返回的 session-id
	Token          string // session信息的唯一token
	DtlsSid        string // dtls协议的 session_id
	MacAddr        string // 客户端mac地址
	UniqueIdGlobal string // 客户端唯一标示
	Username       string // 用户名
	Group          string
	AuthStep       string
	AuthPass       string

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
		if base.Cfg.SessionTimeout == 0 {
			return
		}
		timeout := time.Duration(base.Cfg.SessionTimeout) * time.Second
		tick := time.NewTicker(time.Second * 60)
		for range tick.C {
			sessMux.Lock()
			t := time.Now()
			for k, v := range sessions {
				v.mux.Lock()
				if !v.IsActive {
					if t.Sub(v.LastLogin) > timeout {
						delete(sessions, k)
					}
				}
				v.mux.Unlock()
			}
			sessMux.Unlock()
		}
	}()
}

func GenToken() string {
	// 生成32位的 token
	btoken := make([]byte, 32)
	rand.Read(btoken)
	return fmt.Sprintf("%x", btoken)
}

func NewSession(token string) *Session {
	if token == "" {
		btoken := make([]byte, 32)
		rand.Read(btoken)
		token = fmt.Sprintf("%x", btoken)
	}

	// 生成 dtlsn session_id
	dtlsid := make([]byte, 32)
	rand.Read(dtlsid)

	sess := &Session{
		Sid:       fmt.Sprintf("%d", time.Now().Unix()),
		Token:     token,
		DtlsSid:   fmt.Sprintf("%x", dtlsid),
		LastLogin: time.Now(),
	}

	sessMux.Lock()
	sessions[token] = sess
	sessMux.Unlock()
	return sess
}

func (s *Session) NewConn() *ConnSession {
	s.mux.Lock()
	active := s.IsActive
	macAddr := s.MacAddr
	username := s.Username
	s.mux.Unlock()
	if active {
		s.CSess.Close()
	}

	limit := LimitClient(username, false)
	if !limit {
		return nil
	}
	// 获取客户端mac地址
	macHw, err := net.ParseMAC(macAddr)
	if err != nil {
		sum := md5.Sum([]byte(s.UniqueIdGlobal))
		macHw = sum[0:5] // 5个byte
		macHw = append([]byte{0x02}, macHw...)
		macAddr = macHw.String()
	}
	ip := AcquireIp(username, macAddr)
	if ip == nil {
		LimitClient(username, true)
		return nil
	}

	cSess := &ConnSession{
		Sess:       s,
		MacHw:      macHw,
		IpAddr:     ip,
		closeOnce:  sync.Once{},
		CloseChan:  make(chan struct{}),
		PayloadIn:  make(chan *Payload),
		PayloadOut: make(chan *Payload),
	}

	// 查询group信息
	group := &dbdata.Group{}
	err = dbdata.One("Name", s.Group, group)
	if err != nil {
		base.Error(err)
		cSess.Close()
		return nil
	}
	cSess.Group = group
	if group.Bandwidth > 0 {
		// 限流设置
		cSess.Limit = NewLimitRater(group.Bandwidth, group.Bandwidth)
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
		base.Info("closeOnce:", cs.IpAddr)
		cs.Sess.mux.Lock()
		defer cs.Sess.mux.Unlock()

		close(cs.CloseChan)
		cs.Sess.IsActive = false
		cs.Sess.LastLogin = time.Now()
		cs.Sess.CSess = nil

		ReleaseIp(cs.IpAddr, cs.Sess.MacAddr)
		LimitClient(cs.Sess.Username, true)
	})
}

const BandwidthPeriodSec = 2 // 流量速率统计周期(秒)

func (cs *ConnSession) ratePeriod() {
	tick := time.NewTicker(time.Second * BandwidthPeriodSec)
	defer tick.Stop()

	for range tick.C {
		select {
		case <-cs.CloseChan:
			return
		default:
		}

		// 实时流量清零
		rtUp := atomic.SwapUint32(&cs.BandwidthUp, 0)
		rtDown := atomic.SwapUint32(&cs.BandwidthDown, 0)
		// 设置上一周期每秒的流量
		atomic.SwapUint32(&cs.BandwidthUpPeriod, rtUp/BandwidthPeriodSec)
		atomic.SwapUint32(&cs.BandwidthDownPeriod, rtDown/BandwidthPeriodSec)
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

func (cs *ConnSession) SetTunName(name string) {
	cs.Sess.mux.Lock()
	defer cs.Sess.mux.Unlock()
	cs.TunName = name
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

	return Token2Sess(token)
}

func Token2Sess(token string) *Session {
	sessMux.Lock()
	defer sessMux.Unlock()
	return sessions[token]
}

func Dtls2Sess(dtlsid []byte) *Session {
	return nil
}

func DelSess(token string) {
	// sessions.Delete(token)
}

func CloseSess(token string) {
	sessMux.Lock()
	defer sessMux.Unlock()
	sess, ok := sessions[token]
	if !ok {
		return
	}

	delete(sessions, token)
	sess.CSess.Close()
}

func CloseCSess(token string) {
	sessMux.Lock()
	defer sessMux.Unlock()
	sess, ok := sessions[token]
	if !ok {
		return
	}

	sess.CSess.Close()
}

func DelSessByStoken(stoken string) {
	stoken = strings.TrimSpace(stoken)
	sarr := strings.Split(stoken, "@")
	token := sarr[1]
	sessMux.Lock()
	delete(sessions, token)
	sessMux.Unlock()
}
