package handler

import (
	"testing"
)

// go test -bench=. -benchmem

// 去除数据头
func BenchmarkHeaderCopy(b *testing.B) {
	l := 1500
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		pl := getPayload()
		// 初始化数据
		pl.Data = pl.Data[:l]

		b.StartTimer()
		dataLen := l - 8
		copy(pl.Data, pl.Data[8:8+dataLen])
		// 更新切片长度
		pl.Data = pl.Data[:dataLen]
		b.StopTimer()

		putPayload(pl)
	}
}

func BenchmarkHeaderAppend(b *testing.B) {
	l := 1500
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		pl := getPayload()
		// 初始化数据
		pl.Data = pl.Data[:l]

		b.StartTimer()
		dataLen := l - 8
		pl.Data = append(pl.Data[:0], pl.Data[:8+dataLen]...)
		b.StopTimer()

		putPayload(pl)
	}
}
