package handler

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/bjdgyc/anylink/common"
)

var (
	sessMux  = sync.Mutex{}
	sessions = make(map[string]*Session) // session_token -> SessUser
)

type Session struct {
	Sid     string // auth返回的 session-id
	Token   string // session信息的唯一token
	DtlsSid string // dtls协议的 session_id
	MacAddr string // 客户端mac地址

	// 开启link需要设置的参数
	MasterSecret string // dtls协议的 master_secret
	NetIp        net.IP // 分配的ip地址
	UserName     string // 用户名
	RemoteAddr   string
	Mtu          string
	TunName      string
	IsActive     bool
	LastLogin    time.Time
	closeOnce    sync.Once
	Closed       chan struct{}
	PayloadIn    chan *Payload
	PayloadOut   chan *Payload
}

func init() {
	rand.Seed(time.Now().UnixNano())

	// 检测过期的session
	go func() {
		if common.ServerCfg.SessionTimeout == 0 {
			return
		}
		timeout := time.Duration(common.ServerCfg.SessionTimeout) * time.Second
		tick := time.Tick(time.Second * 30)
		for range tick {
			t := time.Now()
			sessMux.Lock()
			for k, v := range sessions {
				if v.IsActive == true {
					continue
				}
				if t.Sub(v.LastLogin) > timeout {
					delete(sessions, k)
				}
			}
			sessMux.Unlock()
		}
	}()
}

func NewSession() *Session {
	// 生成32位的 token
	btoken := make([]byte, 32)
	rand.Read(btoken)

	// 生成 dtls session_id
	dtlsid := make([]byte, 32)
	rand.Read(dtlsid)

	token := fmt.Sprintf("%x", btoken)
	sess := &Session{
		Sid:       fmt.Sprintf("%d", time.Now().Unix()),
		Token:     token,
		DtlsSid:   fmt.Sprintf("%x", dtlsid),
		LastLogin: time.Now(),
	}
	sessMux.Lock()
	defer sessMux.Unlock()
	sessions[token] = sess
	return sess
}

func (s *Session) StartLink() {
	limit := common.LimitClient(s.UserName, false)
	if limit == false {
		s.NetIp = nil
		return
	}
	s.NetIp = common.AcquireIp(s.MacAddr)
	s.IsActive = true
	s.closeOnce = sync.Once{}
	s.Closed = make(chan struct{})
	s.PayloadIn = make(chan *Payload)
	s.PayloadOut = make(chan *Payload)
}

func (s *Session) Close() {
	s.closeOnce.Do(func() {
		log.Println("closeOnce")
		close(s.Closed)
		s.IsActive = false
		s.LastLogin = time.Now()
		common.ReleaseIp(s.NetIp, s.MacAddr)
		common.LimitClient(s.UserName, true)
	})
}

func SToken2Sess(stoken string) *Session {
	stoken = strings.TrimSpace(stoken)
	sarr := strings.Split(stoken, "@")
	token := sarr[1]
	sessMux.Lock()
	defer sessMux.Unlock()
	if sess, ok := sessions[token]; ok {
		return sess
	}

	return nil
}

func Dtls2Sess(dtlsid []byte) *Session {
	return nil
}

func DelSess(token string) {
	delete(sessions, token)
}

func DelSessByStoken(stoken string) {
	stoken = strings.TrimSpace(stoken)
	sarr := strings.Split(stoken, "@")
	token := sarr[1]
	sessMux.Lock()
	defer sessMux.Unlock()
	delete(sessions, token)
}
