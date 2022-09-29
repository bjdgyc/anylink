package handler

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"time"

	"github.com/bjdgyc/anylink/base"
	"github.com/bjdgyc/anylink/dbdata"
	"github.com/bjdgyc/anylink/pkg/utils"
	"github.com/bjdgyc/anylink/sessdata"
	"github.com/songgao/water/waterutil"
)

const (
	acc_proto_udp = iota + 1
	acc_proto_tcp
	acc_proto_https
	acc_proto_http
)

// 保存批量的审计日志
type LogBatch struct {
	Logs []dbdata.AccessAudit
}

// 日志池
type LogSink struct {
	logChan        chan dbdata.AccessAudit
	autoCommitChan chan *LogBatch // 超时通知
}

var logAuditSink *LogSink

// 写入日志通道
func logAuditWrite(aa dbdata.AccessAudit) {
	logAuditSink.logChan <- aa
}

// 批量写入数据表
func logAuditBatch() {
	if base.Cfg.AuditInterval < 0 {
		return
	}
	logAuditSink = &LogSink{
		logChan:        make(chan dbdata.AccessAudit, 1000),
		autoCommitChan: make(chan *LogBatch, 10),
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
func logAudit(cSess *sessdata.ConnSession, pl *sessdata.Payload) {
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
	ipPort := waterutil.IPv4DestinationPort(pl.Data)

	b := getByte51()
	key := *b
	copy(key[:16], ipSrc)
	copy(key[16:32], ipDst)
	binary.BigEndian.PutUint16(key[32:34], ipPort)

	info := ""
	if ipProto == waterutil.TCP {
		plData := waterutil.IPv4Payload(pl.Data)
		if len(plData) < 14 {
			return
		}
		flags := plData[13]
		switch flags {
		case flags & 0x20:
			// base.Debug("URG "+info, "#", str)
			return
		case flags & 0x10:
			// base.Debug("ACK  ", ipSrc, "#", ipDst, "#", ipPort)
			return
		case flags & 0x08:
			// base.Debug("PSH  "+info, "#", str)
			return
		case flags & 0x04:
			// base.Debug("RST "+info, "#", str)
			return
		case flags & 0x02:
			// base.Debug("SYNC "+info, "#", str)
			return
		case flags & 0x01:
			// base.Debug("FIN "+info, "#", str)
			return
		default:
			accessProto, info = onTCP(plData)
		}
	}
	key[34] = byte(accessProto)
	if info != "" {
		md5Sum := md5.Sum([]byte(info))
		copy(key[35:51], hex.EncodeToString(md5Sum[:]))
	}
	s := utils.BytesToString(key)
	nu := utils.NowSec().Unix()

	// 判断已经存在，并且没有过期
	v, ok := cSess.IpAuditMap.Get(s)
	if ok && nu-v.(int64) < int64(base.Cfg.AuditInterval) {
		// 回收byte对象
		putByte51(b)
		return
	}

	cSess.IpAuditMap.Set(s, nu)

	audit := dbdata.AccessAudit{
		Username:    cSess.Sess.Username,
		Protocol:    uint8(ipProto),
		Src:         ipSrc.String(),
		Dst:         ipDst.String(),
		DstPort:     ipPort,
		CreatedAt:   utils.NowSec(),
		AccessProto: accessProto,
		Info:        info,
	}
	logAuditWrite(audit)
}
