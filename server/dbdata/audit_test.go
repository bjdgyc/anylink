package dbdata

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSearchAudit(t *testing.T) {
	ast := assert.New(t)

	preIpData()
	defer closeIpdata()

	currDateVal := "2022-07-24 00:00:00"
	CreatedAt, _ := time.ParseInLocation("2006-01-02 15:04:05", currDateVal, time.Local)

	dataTest := AccessAudit{
		Username:    "Test",
		Protocol:    6,
		Src:         "10.10.1.5",
		SrcPort:     0,
		Dst:         "172.217.160.68",
		DstPort:     80,
		AccessProto: 4,
		Info:        "www.google.com",
		CreatedAt:   CreatedAt,
	}
	err := Add(dataTest)
	ast.Nil(err)

	var datas []AccessAudit
	searchFormat := `{"username": "%s", "src":"%s", "dst": "%s", "dst_port":"%d","access_proto":"%d","info":"%s","date":["%s","%s"]}`
	search := fmt.Sprintf(searchFormat, dataTest.Username, dataTest.Src, dataTest.Dst, dataTest.DstPort, dataTest.AccessProto, dataTest.Info, currDateVal, currDateVal)

	session := GetAuditSession(search)
	count, _ := FindAndCount(session, &datas, PageSize, 0)
	ast.Equal(count, int64(1))
	ast.Equal(datas[0].Username, dataTest.Username)
	ast.Equal(datas[0].Dst, dataTest.Dst)
}
