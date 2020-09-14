package sessdata

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type A struct {
	Id   int
	Name string
	Age  int
	Addr string
}

type B struct {
	IdB   int
	NameB string
	Age   int
	Addr  string
}

func TestCopyStruct(t *testing.T) {
	assert := assert.New(t)
	a := A{
		Id:   1,
		Name: "bob",
		Age:  15,
		Addr: "American",
	}
	b := B{}
	err := CopyStruct(&b, a)
	assert.Nil(err)
	assert.Equal(b.IdB, 0)
	assert.Equal(b.NameB, "")
	assert.Equal(b.Age, 15)
	assert.Equal(b.Addr, "American")
}
