package sessdata

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/bjdgyc/anylink/base"
)

// func TestCheckUser(t *testing.T) {
// 	user["user1"] = User{Password: "7c4a8d09ca3762af61e59520943dc26494f8941b"}
// 	user["user2"] = User{Password: "7c4a8d09ca3762af61e59520943dc26494f8941c"}
//
// 	var res bool
// 	res = CheckUser("user1", "123456", "")
// 	AssertTrue(t, res == true)
//
// 	res = CheckUser("user2", "123457", "")
// 	AssertTrue(t, res == false)
// }

func TestLimitClient(t *testing.T) {
	assert := assert.New(t)
	base.Cfg.MaxClient = 2
	base.Cfg.MaxUserClient = 1

	res1 := LimitClient("user1", false)
	res2 := LimitClient("user1", false)
	res3 := LimitClient("user2", false)
	res4 := LimitClient("user3", false)
	res5 := LimitClient("user1", true)

	assert.True(res1)
	assert.False(res2)
	assert.True(res3)
	assert.False(res4)
	assert.True(res5)

}

func TestLimitWait(t *testing.T) {
	assert := assert.New(t)
	limit := NewLimitRater(1, 2)
	err := limit.Wait(2)
	assert.Nil(err)
	start := time.Now()
	err = limit.Wait(2)
	assert.Nil(err)
	err = limit.Wait(1)
	assert.Nil(err)
	end := time.Now()
	sub := end.Sub(start)
	assert.Equal(3, int(sub.Seconds()))
}
