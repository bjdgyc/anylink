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
	logAuditSink *LogSink
	auditPayload *AuditPayload
)

// 审计日志
type AuditPayload struct {
	Pool       *grpool.Pool
	IpAuditMap utils.IMaps
}

// 保存批量的审计日志
type LogBatch struct {
	Logs []dbdata.AccessAudit
}

// 批量日志池
type LogSink struct {
	logChan        chan dbdata.AccessAudit
	autoCommitChan chan *LogBatch // 超时通知
}

func (p *AuditPayload) Add(userName string, pl *sessdata.Payload) {
	bPlData := getByteFull()
	copy(*bPlData, pl.Data)
	p.Pool.JobQueue <- func() {
		logAudit(userName, bPlData)
	}
}

// 开启批量写入数据功能
func logAuditBatch() {
	if base.Cfg.AuditInterval < 0 {
		return
	}
	logAuditSink = &LogSink{
		logChan:        make(chan dbdata.AccessAudit, 5000),
		autoCommitChan: make(chan *LogBatch, 10),
	}
	auditPayload = &AuditPayload{
		Pool:       grpool.NewPool(10, 500),
		IpAuditMap: utils.NewMap("cmap", 0),
	}
	var (
		limit        = 100 // 超过上限批量写入数据表
		logAudit     dbdata.AccessAudit
		logBatch     *LogBatch
		commitTimer  *time.Timer // 超时自动提交
		timeOutBatch *LogBatch
	)

	for {
		select {
		case logAudit = <-logAuditSink.logChan:
			if logBatch == nil {
				logBatch = &LogBatch{}
				commitTimer = time.AfterFunc(
					1*time.Second, func(logBatch *LogBatch) func() {
						return func() {
							logAuditSink.autoCommitChan <- logBatch
						}
					}(logBatch),
				)
			}
			logBatch.Logs = append(logBatch.Logs, logAudit)
			if len(logBatch.Logs) >= limit {
				commitTimer.Stop()
				_ = dbdata.AddBatch(logBatch.Logs)
				logBatch = nil
			}
		case timeOutBatch = <-logAuditSink.autoCommitChan:
			if timeOutBatch != logBatch {
				continue
			}
			if logBatch != nil {
				_ = dbdata.AddBatch(logBatch.Logs)
			}
			logBatch = nil
		}
	}
}

// 解析IP包的数据
func logAudit(userName string, bPlData *[]byte) {
	defer putByte(bPlData)

	plData := *bPlData
	ipProto := waterutil.IPv4Protocol(plData)
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

	ipSrc := waterutil.IPv4Source(plData)
	ipDst := waterutil.IPv4Destination(plData)
	ipPort := waterutil.IPv4DestinationPort(plData)

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
		tcpPlData := waterutil.IPv4Payload(plData)
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
	logAuditSink.logChan <- audit
}
