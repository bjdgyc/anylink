package handler

import (
	"encoding/xml"
	"os/exec"

	"github.com/bjdgyc/anylink/base"
)

const BufferSize = 2048

type ClientRequest struct {
	XMLName              xml.Name       `xml:"config-auth"`
	Client               string         `xml:"client,attr"`                 // 一般都是 vpn
	Type                 string         `xml:"type,attr"`                   // 请求类型 init logout auth-reply
	AggregateAuthVersion string         `xml:"aggregate-auth-version,attr"` // 一般都是 2
	Version              string         `xml:"version"`                     // 客户端版本号
	GroupAccess          string         `xml:"group-access"`                // 请求的地址
	GroupSelect          string         `xml:"group-select"`                // 选择的组名
	RemoteAddr           string         `xml:"remote_addr"`
	UserAgent            string         `xml:"user_agent"`
	SessionId            string         `xml:"session-id"`
	SessionToken         string         `xml:"session-token"`
	Auth                 auth           `xml:"auth"`
	DeviceId             deviceId       `xml:"device-id"`
	MacAddressList       macAddressList `xml:"mac-address-list"`
}

type auth struct {
	Username          string `xml:"username"`
	Password          string `xml:"password"`
	OtpSecret         string `xml:"otp_secret"`
	SecondaryPassword string `xml:"secondary_password"`
}

type deviceId struct {
	ComputerName    string `xml:"computer-name,attr"`
	DeviceType      string `xml:"device-type,attr"`
	PlatformVersion string `xml:"platform-version,attr"`
	UniqueId        string `xml:"unique-id,attr"`
	UniqueIdGlobal  string `xml:"unique-id-global,attr"`
}

type macAddressList struct {
	MacAddress string `xml:"mac-address"`
}

func execCmd(cmdStrs []string) error {
	for _, cmdStr := range cmdStrs {
		cmd := exec.Command("sh", "-c", cmdStr)
		b, err := cmd.CombinedOutput()
		if err != nil {
			base.Error(cmdStr, string(b))
			return err
		}
	}
	return nil
}
