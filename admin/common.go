package admin

import (
	"crypto/tls"
	"errors"
	"time"

	"github.com/bjdgyc/anylink/base"
	"github.com/bjdgyc/anylink/dbdata"
	"github.com/dgrijalva/jwt-go"
	"github.com/mojocn/base64Captcha"
	mail "github.com/xhit/go-simple-mail/v2"
)

func SetJwtData(data map[string]interface{}, expiresAt int64) (string, error) {
	jwtData := jwt.MapClaims{"exp": expiresAt}
	for k, v := range data {
		jwtData[k] = v
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtData)

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(base.Cfg.JwtSecret))
	return tokenString, err
}

func GetJwtData(jwtToken string) (map[string]interface{}, error) {
	token, err := jwt.Parse(jwtToken, func(token *jwt.Token) (interface{}, error) {
		// since we only use the one private key to sign the tokens,
		// we also only use its public counter part to verify
		return []byte(base.Cfg.JwtSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("data is parse err")
	}

	return claims, nil
}

func createCaptcha() {
	var store = base64Captcha.DefaultMemStore
	var driver base64Captcha.Driver
	driverString := &base64Captcha.DriverString{}
	driver = driverString.ConvertFonts()
	c := base64Captcha.NewCaptcha(driver, store)
	_ = c
}

func SendMail(subject, to, htmlBody string) error {

	dataSmtp := &dbdata.SettingSmtp{}
	err := dbdata.SettingGet(dataSmtp)
	if err != nil {
		base.Error(err)
		return err
	}

	server := mail.NewSMTPClient()

	// SMTP Server
	server.Host = dataSmtp.Host
	server.Port = dataSmtp.Port
	server.Username = dataSmtp.Username
	server.Password = dataSmtp.Password
	if dataSmtp.UseSSl {
		server.Encryption = mail.EncryptionSSL
	}

	// Since v2.3.0 you can specified authentication type:
	// - PLAIN (default)
	// - LOGIN
	// - CRAM-MD5
	server.Authentication = mail.AuthPlain

	// Variable to keep alive connection
	server.KeepAlive = false

	// Timeout for connect to SMTP Server
	server.ConnectTimeout = 10 * time.Second

	// Timeout for send the data and wait respond
	server.SendTimeout = 10 * time.Second

	// Set TLSConfig to provide custom TLS configuration. For example,
	// to skip TLS verification (useful for testing):
	server.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	// SMTP client
	smtpClient, err := server.Connect()

	if err != nil {
		base.Error(err)
		return err
	}

	// New email simple html with inline and CC
	email := mail.NewMSG()
	email.SetFrom(dataSmtp.From).
		AddTo(to).
		SetSubject(subject)

	email.SetBody(mail.TextHTML, htmlBody)

	// Call Send and pass the client
	err = email.Send(smtpClient)
	if err != nil {
		base.Error(err)
	}

	return err
}
