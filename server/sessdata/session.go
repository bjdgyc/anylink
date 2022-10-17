package sessdata

import (
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
	"github.com/bjdgyc/anylink/pkg/utils"
	mapset "github.com/deckarep/golang-set"
	"github.com/ivpusic/grpool"
	atomic2 "go.uber.org/atomic"
)

var (
	// session_token -> SessUser
	sessions = make(map[string]*Session)
	// dtlsId -> session_token
	dtlsIds = make(map[string]string)
	sessMux sync.RWMutex
)

// 连接sess
type ConnSession struct {
	Sess                *Session
	MasterSecret        string // dtls协议的 master_secret
	IpAddr              net.IP // 分配的ip地址
	LocalIp             net.IP
	MacHw               net.HardwareAddr // 客户端mac地址,从Session取出
	Username            string
	RemoteAddr          string
	Mtu                 int
	IfName              string
	Client              string // 客户端  mobile pc
	CstpDpd             int
	Group               *dbdata.Group
	Limit               *LimitRater
	BandwidthUp         atomic2.Uint32 // 使用上行带宽 Byte
	BandwidthDown       atomic2.Uint32 // 使用下行带宽 Byte
	BandwidthUpPeriod   atomic2.Uint32 // 前一周期的总量
	BandwidthDownPeriod atomic2.Uint32
	BandwidthUpAll      atomic2.Uint64 // 使用上行带宽总量
	BandwidthDownAll    atomic2.Uint64 // 使用下行带宽总量
	closeOnce           sync.Once
	CloseChan           chan struct{}
	PayloadIn           chan *Payload
	PayloadOutCstp      chan *Payload // Cstp的数据
	PayloadOutDtls      chan *Payload // Dtls的数据
	IpAuditMap          utils.IMaps   // 审计的ip数据
	IpAuditPool         *grpool.Pool  // 审计的IP包解析池
	// dSess *DtlsSession
	dSess *atomic.Value
}

type DtlsSession struct {
	isActive  int32
	CloseChan chan struct{}
	closeOnce sync.Once
	IpAddr    net.IP
}

type Session struct {
	mux            sync.RWMutex
	Sid            string // auth返回的 session-id
	Token          string // session信息的唯一token
	DtlsSid        string // dtls协议的 session_id
	MacAddr        string // 客户端mac地址
	UniqueIdGlobal string // 客户端唯一标示
	MacHw          net.HardwareAddr
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

// 状态为过期的用户踢下线
func CloseUserLimittimeSession() {
	s := mapset.NewSetFromSlice(dbdata.CheckUserlimittime())
	for _, v := range sessions {
		if v.IsActive && s.Contains(v.Username) {
			CloseSess(v.Token)
		}
	}
}

func GenToken() string {
	// 生成32位的 token
	bToken := make([]byte, 32)
	rand.Read(bToken)
	return fmt.Sprintf("%x", bToken)
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
	dtlsIds[sess.DtlsSid] = token
	sessMux.Unlock()
	return sess
}

func (s *Session) NewConn() *ConnSession {
	s.mux.RLock()
	active := s.IsActive
	macAddr := s.MacAddr
	macHw := s.MacHw
	username := s.Username
	s.mux.RUnlock()
	if active {
		s.CSess.Close()
	}

	limit := LimitClient(username, false)
	if !limit {
		return nil
	}
	ip := AcquireIp(username, macAddr)
	if ip == nil {
		LimitClient(username, true)
		return nil
	}

	// 查询group信息
	group := &dbdata.Group{}
	err := dbdata.One("Name", s.Group, group)
	if err != nil {
		base.Error(err)
		return nil
	}

	cSess := &ConnSession{
		Sess:           s,
		MacHw:          macHw,
		Username:       username,
		IpAddr:         ip,
		closeOnce:      sync.Once{},
		CloseChan:      make(chan struct{}),
		PayloadIn:      make(chan *Payload, 64),
		PayloadOutCstp: make(chan *Payload, 64),
		PayloadOutDtls: make(chan *Payload, 64),
		dSess:          &atomic.Value{},
	}

	// ip 审计
	if base.Cfg.AuditInterval >= 0 {
		cSess.IpAuditMap = utils.NewMap("cmap", 0)
		cSess.IpAuditPool = grpool.NewPool(1, 600)
	}

	dSess := &DtlsSession{
		isActive: -1,
	}
	cSess.dSess.Store(dSess)

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

		dSess := cs.GetDtlsSession()
		if dSess != nil {
			dSess.Close()
		}

		ReleaseIp(cs.IpAddr, cs.Sess.MacAddr)
		LimitClient(cs.Username, true)
	})
}

// 创建dtls链接
func (cs *ConnSession) NewDtlsConn() *DtlsSession {
	ds := cs.dSess.Load().(*DtlsSession)
	isActive := atomic.LoadInt32(&ds.isActive)
	if isActive > 0 {
		// 判断原有连接存在，不进行创建
		return nil
	}

	dSess := &DtlsSession{
		isActive:  1,
		CloseChan: make(chan struct{}),
		closeOnce: sync.Once{},
		IpAddr:    cs.IpAddr,
	}
	cs.dSess.Store(dSess)
	return dSess
}

// 关闭dtls链接
func (ds *DtlsSession) Close() {
	ds.closeOnce.Do(func() {
		base.Info("closeOnce dtls:", ds.IpAddr)

		atomic.StoreInt32(&ds.isActive, -1)
		close(ds.CloseChan)
	})
}

func (cs *ConnSession) GetDtlsSession() *DtlsSession {
	ds := cs.dSess.Load().(*DtlsSession)
	isActive := atomic.LoadInt32(&ds.isActive)
	if isActive > 0 {
		return ds
	}
	return nil
}

const BandwidthPeriodSec = 10 // 流量速率统计周期(秒)

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
		rtUp := cs.BandwidthUp.Swap(0)
		rtDown := cs.BandwidthDown.Swap(0)
		// 设置上一周期每秒的流量
		cs.BandwidthUpPeriod.Swap(rtUp / BandwidthPeriodSec)
		cs.BandwidthDownPeriod.Swap(rtDown / BandwidthPeriodSec)
		// 累加所有流量
		cs.BandwidthUpAll.Add(uint64(rtUp))
		cs.BandwidthDownAll.Add(uint64(rtDown))
	}
}

var MaxMtu = 1460

func (cs *ConnSession) SetMtu(mtu string) {
	if base.Cfg.Mtu > 0 {
		MaxMtu = base.Cfg.Mtu
	}
	cs.Mtu = MaxMtu

	mi, err := strconv.Atoi(mtu)
	if err != nil || mi < 100 {
		return
	}

	if mi < MaxMtu {
		cs.Mtu = mi
	}
}

func (cs *ConnSession) SetIfName(name string) {
	cs.Sess.mux.Lock()
	defer cs.Sess.mux.Unlock()
	cs.IfName = name
}

func (cs *ConnSession) RateLimit(byt int, isUp bool) error {
	if isUp {
		cs.BandwidthUp.Add(uint32(byt))
		return nil
	}
	// 只对下行速率限制
	cs.BandwidthDown.Add(uint32(byt))
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
	sessMux.RLock()
	defer sessMux.RUnlock()
	return sessions[token]
}

func Dtls2Sess(did string) *Session {
	sessMux.RLock()
	defer sessMux.RUnlock()
	token := dtlsIds[did]
	return sessions[token]
}

func Dtls2MasterSecret(did string) string {
	sessMux.RLock()
	token := dtlsIds[did]
	sess := sessions[token]
	sessMux.RUnlock()

	if sess == nil {
		return ""
	}

	sess.mux.RLock()
	defer sess.mux.RUnlock()
	if sess.CSess == nil {
		return ""
	}
	return sess.CSess.MasterSecret
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
	sessMux.RLock()
	defer sessMux.RUnlock()
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
