package handler

import (
	"sync"

	"github.com/bjdgyc/anylink/sessdata"
)

var plPool = sync.Pool{
	New: func() interface{} {
		b := make([]byte, 0, BufferSize)
		pl := sessdata.Payload{
			Data: &b,
		}
		// fmt.Println("plPool-init", len(pl.Data), cap(pl.Data))
		return &pl
	},
}

func getPayload() *sessdata.Payload {
	pl := plPool.Get().(*sessdata.Payload)
	return pl
}

func putPayload(pl *sessdata.Payload) {
	pl.LType = 0
	pl.PType = 0
	*pl.Data = (*pl.Data)[:0]
	plPool.Put(pl)
}

var bytePool = sync.Pool{
	New: func() interface{} {
		b := make([]byte, 0, BufferSize)
		// fmt.Println("bytePool-init")
		return &b
	},
}

func getByteZero() *[]byte {
	b := bytePool.Get().(*[]byte)
	return b
}

func getByteFull() *[]byte {
	b := bytePool.Get().(*[]byte)
	*b = (*b)[:BufferSize]
	return b
}
func putByte(b *[]byte) {
	*b = (*b)[:0]
	bytePool.Put(b)
}
