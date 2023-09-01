package handler

import (
	"crypto/md5"
	"encoding/binary"
	"time"

	"github.com/bjdgyc/anylink/base"
	"github.com/bjdgyc/anylink/dbdata"
	"github.com/bjdgyc/anylink/pkg/utils"
	"github.com/bjdgyc/anylink/sessdata"
	"github.com/ivpusic/grpool"
	"github.com/songgao/water/waterutil"
)

const (
	acc_proto_udp = iota + 1
	acc_proto_tcp
	acc_proto_https
	acc_proto_http
)

var (
	auditPayload *AuditPayload
	logBatch     *LogBatch
)

// 分析审计日志
type AuditPayload struct {
	Pool       *grpool.Pool
	IpAuditMap utils.IMaps
}

// 保存审计日志
type LogBatch struct {
	Logs    []dbdata.AccessAudit
	LogChan chan dbdata.AccessAudit
}

// 异步写入pool
func (p *AuditPayload) Add(userName string, pl *sessdata.Payload) {
	select {
	case p.Pool.JobQueue <- func() {
		logAudit(userName, pl)
	}:
	default:
		putPayload(pl)
		base.Error("AccessAudit: AuditPayload channel is full")
	}
}

// 数据落盘
func (l *LogBatch) Write() {
	if len(l.Logs) == 0 {
		return
	}
	_ = dbdata.AddBatch(l.Logs)
	l.Reset()
}

// 清空数据
func (l *LogBatch) Reset() {
	l.Logs = []dbdata.AccessAudit{}
}

// 开启批量写入数据功能
func logAuditBatch() {
	if base.Cfg.AuditInterval < 0 {
		return
	}
	auditPayload = &AuditPayload{
		Pool:       grpool.NewPool(10, 10240),
		IpAuditMap: utils.NewMap("cmap", 0),
	}
	logBatch = &LogBatch{
		LogChan: make(chan dbdata.AccessAudit, 10240),
	}
	var (
		limit       = 100 // 超过上限批量写入数据表
		outTime     = time.NewTimer(time.Second)
		accessAudit = dbdata.AccessAudit{}
	)

	for {
		// 重置超时 时间
		outTime.Reset(time.Second * 1)
		select {
		case accessAudit = <-logBatch.LogChan:
			logBatch.Logs = append(logBatch.Logs, accessAudit)
			if len(logBatch.Logs) >= limit {
				if !outTime.Stop() {
					<-outTime.C
				}
				logBatch.Write()
			}
		case <-outTime.C:
			logBatch.Write()
		}
	}
}

// 解析IP包的数据
func logAudit(userName string, pl *sessdata.Payload) {
	defer putPayload(pl)

	if !(pl.LType == sessdata.LTypeIPData && pl.PType == 0x00) {
		return
	}

	ipProto := waterutil.IPv4Protocol(pl.Data)
	// 访问协议
	var accessProto uint8
	// 只统计 tcp和udp 的访问
	switch ipProto {
	case waterutil.TCP:
		accessProto = acc_proto_tcp
	case waterutil.UDP:
		accessProto = acc_proto_udp
	default:
		return
	}

	ipSrc := waterutil.IPv4Source(pl.Data)
	ipDst := waterutil.IPv4Destination(pl.Data)

	// ipPort := waterutil.IPv4DestinationPort(pl.Data)
	// 修复 panic: runtime error: index out of range [2] with length 2
	ipPl := waterutil.IPv4Payload(pl.Data)
	if len(ipPl) < 3 {
		base.Error("ipPl len < 3", pl.Data)
		return
	}
	ipPort := (uint16(ipPl[2]) << 8) | uint16(ipPl[3])

	b := getByte51()
	key := *b
	copy(key[:16], ipSrc)
	copy(key[16:32], ipDst)
	binary.BigEndian.PutUint16(key[32:34], ipPort)
	key[34] = byte(accessProto)
	copy(key[35:51], []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})

	info := ""
	nu := utils.NowSec().Unix()
	if ipProto == waterutil.TCP {
		tcpPlData := waterutil.IPv4Payload(pl.Data)
		// 24 (ACK PSH)
		if len(tcpPlData) < 14 || tcpPlData[13] != 24 {
			return
		}
		accessProto, info = onTCP(tcpPlData)
		// HTTPS or HTTP
		if accessProto != acc_proto_tcp {
			// 提前存储只含ip数据的key, 避免即记录域名又记录一笔IP数据的记录
			ipKey := make([]byte, 51)
			copy(ipKey, key)
			ipS := utils.BytesToString(ipKey)
			auditPayload.IpAuditMap.Set(ipS, nu)

			key[34] = byte(accessProto)
			// 存储含域名的key
			if info != "" {
				md5Sum := md5.Sum([]byte(info))
				copy(key[35:51], md5Sum[:])
			}
		}
	}
	s := utils.BytesToString(key)

	// 判断已经存在，并且没有过期
	v, ok := auditPayload.IpAuditMap.Get(s)
	if ok && nu-v.(int64) < int64(base.Cfg.AuditInterval) {
		// 回收byte对象
		putByte51(b)
		return
	}

	auditPayload.IpAuditMap.Set(s, nu)

	audit := dbdata.AccessAudit{
		Username:    userName,
		Protocol:    uint8(ipProto),
		Src:         ipSrc.String(),
		Dst:         ipDst.String(),
		DstPort:     ipPort,
		CreatedAt:   utils.NowSec(),
		AccessProto: accessProto,
		Info:        info,
	}

	select {
	case logBatch.LogChan <- audit:
	default:
		base.Error("AccessAudit: LogChan channel is full")
		return
	}
}
