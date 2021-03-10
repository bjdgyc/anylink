package dbdata

import (
	"testing"

	"github.com/bjdgyc/anylink/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestGetGroupNames(t *testing.T) {
	ast := assert.New(t)

	preIpData()
	defer closeIpdata()

	// 添加 group
	g1 := Group{Name: "g1", ClientDns: []ValData{{Val: "114.114.114.114"}}}
	err := SetGroup(&g1)
	ast.Nil(err)
	g2 := Group{Name: "g2", ClientDns: []ValData{{Val: "114.114.114.114"}}}
	err = SetGroup(&g2)
	ast.Nil(err)
	g3 := Group{Name: "g3", ClientDns: []ValData{{Val: "114.114.114.114"}}}
	err = SetGroup(&g3)
	ast.Nil(err)

	// 判断所有数据
	gAll := []string{"g1", "g2", "g3"}
	gs := GetGroupNames()
	for _, v := range gs {
		ast.Equal(true, utils.InArrStr(gAll, v))
	}
}
