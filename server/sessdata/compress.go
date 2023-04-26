package sessdata

import (
	"github.com/lanrenwo/lzsgo"
)

type CmpEncoding interface {
	Compress(src []byte, dst []byte) (int, error)
	Uncompress(src []byte, dst []byte) (int, error)
}

type LzsgoCmp struct {
}

func (l LzsgoCmp) Compress(src []byte, dst []byte) (int, error) {
	n, err := lzsgo.Compress(src, dst)
	return n, err
}

func (l LzsgoCmp) Uncompress(src []byte, dst []byte) (int, error) {
	n, err := lzsgo.Uncompress(src, dst)
	return n, err
}

// type Lz4Cmp struct {
// 	c lz4.Compressor
// }

// func (l Lz4Cmp) Compress(src []byte, dst []byte) (int, error) {
// 	return l.c.CompressBlock(src, dst)
// }

// func (l Lz4Cmp) Uncompress(src []byte, dst []byte) (int, error) {
// 	return lz4.UncompressBlock(src, dst)
// }
