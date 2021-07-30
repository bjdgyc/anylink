package handler

import (
	"testing"
)

// go test -bench=. -benchmem

// Strings written to buf
var strs = []string{
	"Lorem ipsum dolor sit amet, consectetur adipiscing elit",
	"sed do eiusmod tempor incididunt ut labore et dolore magna aliqua",
	`Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris
		nisi ut aliquip ex ea commodo consequat.
		Duis aute irure dolor in reprehenderit in voluptate velit esse cillum
		dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident,
		sunt in culpa qui officia deserunt mollit anim id est laborum`,
	"Sed ut perspiciatis",
	"sed quia consequuntur magni dolores eos qui ratione voluptatem sequi nesciunt",
	"Ut enim ad minima veniam, quis nostrum exercitationem ullam corporis suscipit",
	"laboriosam, nisi ut aliquid ex ea commodi consequatur",
	"Quis autem vel eum iure reprehenderit qui in ea voluptate velit esse quam nihil molestiae consequatur",
	"vel illum qui dolorem eum fugiat quo voluptas nulla pariatur",
}

// 去除数据头
func BenchmarkHeaderCopy(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, v := range strs {
			pl := getPayload()
			// 初始化数据
			pl.Data = append(pl.Data[:0], v...)

			dataLen := len(v) - 8
			copy(pl.Data, pl.Data[8:8+dataLen])
			// 更新切片长度
			pl.Data = pl.Data[:dataLen]

			putPayload(pl)
		}
	}
}

func BenchmarkHeaderAppend(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, v := range strs {
			pl := getPayload()
			// 初始化数据
			pl.Data = append(pl.Data[:0], v...)

			dataLen := len(v) - 8
			pl.Data = append(pl.Data[:0], pl.Data[:8+dataLen]...)

			putPayload(pl)
		}
	}
}
