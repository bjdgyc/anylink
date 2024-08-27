package dbdata

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"io"
	"log"
	"net/http"
)

type OtpAuthResult struct {
	User       string `json:"user"`
	TokenValid bool   `json:"token_valid"`
}

func ValidateUserOtp(name string, otp int) (bool, error) {

	v := viper.New()
	v.SetConfigFile("./conf/server.toml")
	if err := v.ReadInConfig(); err != nil {
		panic("config file err:" + err.Error())

	}

	// 验证动态口令
	otpServ := v.Get("otp_server")
	otpAuthUrl := fmt.Sprintf("%s/%s/token/%d", otpServ, name, otp)
	fmt.Println("otpAuthUrl: ", otpAuthUrl)
	resp, err := http.Get(otpAuthUrl)

	if err != nil || resp.StatusCode != 200 {
		log.Fatalf("otp server auth err, user=[%s], token=[%d], httpcode=[%d], err=[%v]", name, otp, resp.StatusCode, err)
		return false, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(resp.Body)

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("io.ReadAll read http response body failed, err=[%v]", err)
		return false, err
	}

	var optAuthResult OtpAuthResult
	err = json.Unmarshal(b, &optAuthResult)

	if err != nil {
		log.Fatalf("unmarshalotp retmsg failed, user=[%s], token=[%d], httpcode=[%d], err=[%v]", name, otp, resp.StatusCode, err)
		return false, err
	}

	return optAuthResult.TokenValid, nil
}
