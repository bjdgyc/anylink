package common

import (
	"testing"
)

func TestCheckUser(t *testing.T) {
	users["user1"] = User{Password: "7c4a8d09ca3762af61e59520943dc26494f8941b"}
	users["user2"] = User{Password: "7c4a8d09ca3762af61e59520943dc26494f8941c"}

	var res bool
	res = CheckUser("user1", "123456", "")
	AssertTrue(t, res == true)

	res = CheckUser("user2", "123457", "")
	AssertTrue(t, res == false)
}

func TestLimitClient(t *testing.T) {
	ServerCfg.MaxClient = 2
	ServerCfg.MaxUserClient = 1

	res1 := LimitClient("user1", false)
	res2 := LimitClient("user1", false)
	res3 := LimitClient("user2", false)
	res4 := LimitClient("user3", false)

	AssertTrue(t, res1 == true)
	AssertTrue(t, res2 == false)
	AssertTrue(t, res3 == true)
	AssertTrue(t, res4 == false)

}
