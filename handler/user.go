package handler

import (
	"crypto/sha1"
	"fmt"
	"os"
	"time"

	"github.com/bjdgyc/anylink/common"
	"github.com/bjdgyc/anylink/dbdata"
	"github.com/xlzd/gotp"
)

func CheckUser(name, pwd, group string) bool {
	return true

	pl := len(pwd)
	if name == "" || pl < 6 {
		return false
	}
	v := &dbdata.User{}
	err := dbdata.Get(dbdata.BucketUser, name, v)
	if err != nil {
		return false
	}
	if !common.InArrStr(v.Group, group) {
		return false
	}
	pass := pwd[:pl-6]
	pwdHash := hashPass(pass)
	if v.Password != pwdHash {
		return false
	}
	otp := pwd[pl-6:]
	totp := gotp.NewDefaultTOTP(v.OtpSecret)
	unix := time.Now().Unix()
	verify := totp.Verify(otp, int(unix))
	if !verify {
		return false
	}
	return true
}

func UserAdd(name, pwd string, group []string) dbdata.User {
	v := dbdata.User{
		Id:        dbdata.NextId(dbdata.BucketUser),
		Username:  name,
		Password:  hashPass(pwd),
		OtpSecret: gotp.RandomSecret(32),
		Group:     group,
		UpdatedAt: time.Now(),
	}
	fmt.Println(v)
	secret := "WHH7WA6POOGGEYVIQYXLZU75QLM7YLUX"
	totp := gotp.NewDefaultTOTP(secret)
	s := totp.ProvisioningUri("bjdtest", "bjdpro")
	fmt.Println(s)

	// qr, _ := qrcode.New(s, qrcode.Medium)
	// a := qr.ToSmallString(false)
	// fmt.Println(a)
	// qr.WriteFile(512, "a.png")

	os.Exit(0)
	return v
}

func hashPass(pwd string) string {
	sum := sha1.Sum([]byte(pwd))
	return fmt.Sprintf("%x", sum)
}
