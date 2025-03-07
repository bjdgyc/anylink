package admin

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/bjdgyc/anylink/base"
	"github.com/bjdgyc/anylink/dbdata"
	"github.com/golang-jwt/jwt/v4"
	"net"
	"net/http"
	"net/smtp"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func GenerateResetToken(userID int) (string, error) {
	fmt.Println(base.Cfg.JwtSecret)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(1 * time.Hour).Unix(),
	})
	return token.SignedString([]byte(base.Cfg.JwtSecret))
}

type CustomClaims struct {
	UserID int `json:"user_id"`
	jwt.StandardClaims
}

func ValidateResetToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&CustomClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(base.Cfg.JwtSecret), nil
		},
	)

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, err
}

// 重置密码函数
func ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token    string `json:"token" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fmt.Println("json 解析失败", err)
		RespError(w, RespInternalErr, "json 解析失败")
		return
	} else {
		// 验证token,并获取UserID
		claims, valid_err := ValidateResetToken(req.Token)
		if valid_err != nil {
			fmt.Println("Token 链接无效或已过期", err)
			RespError(w, RespInternalErr, "Token 链接无效或已过期")
			return
		}
		// 根据验证后的UserId 来更新用户表的密码
		s := &dbdata.User{PinCode: req.Password}
		update_err := dbdata.Update("Id", claims.UserID, s)
		if update_err != nil {
			fmt.Println("更新密码失败", update_err)
			RespError(w, RespInternalErr, "更新密码失败")
			return
		}
		fmt.Println("更新密码成功")
		// 删除重置记录表的数据
		reset := dbdata.PasswordReset{UserId: claims.UserID}
		del_err := dbdata.Del(&reset)

		if del_err != nil {
			fmt.Println("删除记录失败", del_err)
		} else {
			fmt.Println("删除验证记录成功,UserId", claims.UserID)
		}
		RespSucess(w, "密码重置成功")
	}
}

// 发送重置密码请求
func ForgotPassword(w http.ResponseWriter, r *http.Request) {
	// 校验 Content-Type
	if r.Header.Get("Content-Type") != "application/json" {
		RespError(w, RespInternalErr, "仅支持 JSON 格式")
		return
	}
	// 解析json数据
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespError(w, RespInternalErr, "Json 解析失败")
	}
	// 获取用户xorm 数据
	user := &dbdata.User{}
	err := dbdata.One("Email", req.Email, user)

	fmt.Println("获取到的用户ID,", user.Id)
	if user.Id == 0 {
		RespError(w, RespInternalErr, "用户不存在 输入的地址:"+req.Email)
		return
	}
	// 根据用户ID 获取token 数据
	token, err := GenerateResetToken(user.Id)
	if err != nil {
		RespError(w, RespInternalErr, "生成token失败")
		return
	}
	// 插入token 数据到passwordReset 表中
	reset := &dbdata.PasswordReset{}
	reset.UserId = user.Id
	reset.Token = token
	reset.ExpiresAt = int(time.Now().Unix())

	insert_err := dbdata.Add(reset)
	if insert_err != nil {
		fmt.Println("插入token表失败", insert_err)
		RespError(w, RespInternalErr, "插入token表失败")
		return
	}

	// 获取邮箱服务器的配置
	dataSmtp := &dbdata.SettingSmtp{}
	serverConf := &dbdata.SettingOther{}

	mail_err := dbdata.SettingGet(dataSmtp)
	server_err := dbdata.SettingGet(serverConf)
	if server_err != nil {
		fmt.Println("获取服务器配置失败", err)
		RespError(w, RespInternalErr, "获取服务器配置失败,请检查后台对外地址的配置")
		return
	}
	if mail_err != nil {
		fmt.Println("获取邮箱配置失败", err)
		RespError(w, RespInternalErr, "获取邮箱配置失败,请检查邮箱配置")
		return
	}
	// 根据后台配置的对外地址进行解析
	parsedURL, err := url.Parse(serverConf.LinkAddr)
	if err != nil {
		fmt.Println("解析 URL 失败:", err)
		RespError(w, RespInternalErr, "解析URL失败,可能是后台配置的对外地址不符合要求,应该按照https://xxx.test.com or https://xxx.test.com:10443")
		return
	}
	// 提取协议（http 或 https）
	scheme := parsedURL.Scheme
	// 提取主机名（域名）
	hosts := parsedURL.Hostname()
	// 拼接协议和域名,加上后台的地址 就是发送重置链接的地址
	fullURL := fmt.Sprintf("%s://%s%s", scheme, hosts, base.Cfg.AdminAddr)
	// 输出结果
	resetLink := fmt.Sprintf("%s/ui/#resetPassword?token=%s", fullURL, reset.Token)
	mail_user := dataSmtp.Username
	password := dataSmtp.Password
	host := dataSmtp.Host
	port := strconv.Itoa(dataSmtp.Port)
	toUser := req.Email

	message := fmt.Sprintf("这个是vpn账号的重置链接:%s \n密码有效期1小时,超时请重新提交", resetLink)
	if err := SendResetMail(mail_user, password, toUser, "vpn密码重置", message, host, port, false); err != nil {
		fmt.Println("发送失败")
		RespError(w, RespInternalErr, "邮箱发送失败")
		return
	} else {
		fmt.Println("send mail success")
		RespSucess(w, "邮箱发送成功")
		return
	}
}

func SendResetMail(email, password, toEmail, subject, body, host, port string, isHtml bool) (err error) {
	header := make(map[string]string)
	header["From"] = "<" + email + ">"
	header["To"] = toEmail
	header["Subject"] = subject
	if isHtml {
		header["Content-Type"] = "text/html; charset=UTF-8"
	} else {
		header["Content-Type"] = "text/plain; charset=UTF-8"
	}

	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	auth := smtp.PlainAuth(
		"",
		email,
		password,
		host,
	)

	toEmails := strings.Split(toEmail, ";")
	fmt.Println(email, password, toEmails, host, port)
	err = sendMailUsingTLS(
		fmt.Sprintf("%s:%s", host, port),
		auth,
		email,
		toEmails,
		[]byte(message),
	)

	if err != nil {
		fmt.Printf("send_mail_error: %v, %v", toEmails, err)
	}
	return
}

// return a smtp client
func dial(addr string) (*smtp.Client, error) {
	conn, err := tls.Dial("tcp", addr, nil)
	if err != nil {
		fmt.Println("Dialing Error:", err)
		return nil, err
	}
	//分解主机端口字符串
	host, _, _ := net.SplitHostPort(addr)
	return smtp.NewClient(conn, host)
}

func sendMailUsingTLS(addr string, auth smtp.Auth, from string,
	to []string, msg []byte) (err error) {

	//create smtp client
	c, err := dial(addr)
	if err != nil {
		fmt.Println("Create smpt client error:", err)
		return err
	}
	defer c.Close()

	if auth != nil {
		if ok, _ := c.Extension("AUTH"); ok {
			if err = c.Auth(auth); err != nil {
				fmt.Println("Error during AUTH", err)
				return err
			}
		}
	}

	if err = c.Mail(from); err != nil {
		return err
	}

	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			return err
		}
	}

	w, err := c.Data()
	if err != nil {
		return err
	}

	_, err = w.Write(msg)
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return c.Quit()
}
