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

var forgot_interval_time = 1 * 60 // 1分钟的间隔时间（单位：秒）
var reset_interval_time = 30      // 密码重置有效期(单位: 分钟)

func GenerateResetToken(userID int) (string, error) {
	fmt.Println(base.Cfg.JwtSecret)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Duration(reset_interval_time) * time.Minute).Unix(),
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
		RespError(w, RespInternalErr, "json 解析失败")
		return
	} else {
		// 验证token,并获取UserID
		claims, valid_err := ValidateResetToken(req.Token)
		if valid_err != nil {
			msg := fmt.Sprintf("验证失败, 重置链接已过期, 请重新申请。错误信息: %v", valid_err)
			RespError(w, RespInternalErr, msg)
			return
		}
		// 根据验证后的UserId 来更新用户表的密码
		s := &dbdata.User{PinCode: req.Password}
		update_err := dbdata.Update("Id", claims.UserID, s)
		if update_err != nil {
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
		return
	}
	// 获取用户数据
	user := &dbdata.User{}
	err := dbdata.One("Email", req.Email, user)
	if err != nil {
		RespError(w, RespInternalErr, "用户不存在 输入的地址:"+req.Email)
		return
	}
	// 检查上一次请求的时间
	reset := &dbdata.PasswordReset{}
	err = dbdata.One("user_id", user.Id, reset)
	if err != nil {
		if err == dbdata.ErrNotFound {
			base.Info("此账号没有重置记录")
		} else {
			RespError(w, RespInternalErr, "查询重置记录失败", err.Error())
			return
		}
	} else {
		// 检查时间间隔是否足够
		currentTime := int(time.Now().Unix())
		lastRequestTime := reset.LastRequestTime
		if currentTime-lastRequestTime < forgot_interval_time {
			msg := fmt.Sprintf("重复的重置操作,请等待%d秒后进行重置申请", forgot_interval_time)
			RespError(w, RespInternalErr, msg)
			return
		}
	}
	// 生成新的重置令牌
	token, err := GenerateResetToken(user.Id)
	if err != nil {
		RespError(w, RespInternalErr, "生成token失败")
		return
	}
	// 开始更新或插入重置记录
	reset.ExpiresAt = int(time.Now().Add(time.Duration(reset_interval_time) * time.Minute).Unix())
	reset.LastRequestTime = int(time.Now().Unix())
	reset.UserId = user.Id

	if reset.Token == "" {
		// 如果 Token 为空，说明是第一次请求，插入新记录
		reset.Token = token
		err = dbdata.Add(reset)
	} else {
		// 如果 Token 不为空，说明记录已存在，更新记录
		reset.Token = token
		err = dbdata.Update("user_id", reset.UserId, reset)
	}
	if err != nil {
		RespError(w, RespInternalErr, "更新重置记录失败")
		return
	}

	// 获取邮箱服务器的配置
	dataSmtp := &dbdata.SettingSmtp{}
	serverConf := &dbdata.SettingOther{}

	mail_err := dbdata.SettingGet(dataSmtp)
	server_err := dbdata.SettingGet(serverConf)
	if server_err != nil {
		RespError(w, RespInternalErr, "获取服务器配置失败,请检查后台对外地址的配置")
		return
	}
	if mail_err != nil {
		RespError(w, RespInternalErr, "获取邮箱配置失败,请检查邮箱配置")
		return
	}

	// 构建重置链接
	parsedURL, err := url.Parse(serverConf.LinkAddr)
	if err != nil {
		RespError(w, RespInternalErr, "解析URL失败,可能是后台配置的对外地址不符合要求")
		return
	}

	scheme := parsedURL.Scheme
	hosts := parsedURL.Hostname()
	fullURL := fmt.Sprintf("%s://%s%s", scheme, hosts, base.Cfg.AdminAddr)
	resetLink := fmt.Sprintf("%s/ui/#resetPassword?token=%s", fullURL, reset.Token)

	// 发送邮件
	mail_user := dataSmtp.Username
	password := dataSmtp.Password
	host := dataSmtp.Host
	port := strconv.Itoa(dataSmtp.Port)
	toUser := req.Email

	message := fmt.Sprintf("这个是vpn账号的重置链接:%s \n密码有效期%d分钟,超时请重新提交", resetLink, reset_interval_time)
	if err := SendResetMail(mail_user, password, toUser, "vpn密码重置", message, host, port, false); err != nil {
		RespError(w, RespInternalErr, "邮箱发送失败")
		return
	}
	RespSucess(w, "邮箱发送成功")
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

