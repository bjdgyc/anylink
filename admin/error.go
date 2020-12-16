package admin

// 返回码
const (
	RespSuccess       = 0
	RespInternalErr   = 1
	RespTokenErr      = 2
	RespUserOrPassErr = 3
	RespParamErr      = 4
)

var RespMap = map[int]string{
	RespTokenErr:      "客户端TOKEN错误",
	RespUserOrPassErr: "用户名或密码错误",
}
