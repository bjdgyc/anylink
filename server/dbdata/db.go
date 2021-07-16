package dbdata

import (
	"encoding/json"

	"github.com/bjdgyc/anylink/base"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"xorm.io/xorm"
)

var (
	xdb *xorm.Engine
)

func GetXdb() *xorm.Engine {
	return xdb
}

func initDb() {
	var err error
	xdb, err = xorm.NewEngine(base.Cfg.DbType, base.Cfg.DbSource)
	// xdb.ShowSQL(true)
	if err != nil {
		base.Fatal(err)
	}

	// 初始化数据库
	err = xdb.Sync2(&User{}, &Setting{}, &Group{}, &IpMap{})
	if err != nil {
		base.Fatal(err)
	}

	// fmt.Println("s1=============", err)
}

func initData() {
	var (
		err error
	)

	// 判断是否初次使用
	s := &Setting{}
	err = One("name", InstallName, s)
	if err == nil && s.Data == InstallData {
		// 已经安装过
		return
	}

	err = addInitData()
	if err != nil {
		base.Fatal(err)
	}
}

func addInitData() error {
	var (
		err error
	)

	sess := xdb.NewSession()
	defer sess.Close()

	err = sess.Begin()
	if err != nil {
		return err
	}

	// SettingSmtp
	smtp := &SettingSmtp{
		Host:       "127.0.0.1",
		Port:       25,
		From:       "vpn@xx.com",
		Encryption: "None",
	}
	v, _ := json.Marshal(smtp)
	s := &Setting{Name: StructName(smtp), Data: string(v)}
	_, err = sess.InsertOne(s)
	if err != nil {
		return err
	}

	// SettingOther
	other := &SettingOther{
		LinkAddr:    "vpn.xx.com",
		Banner:      "您已接入公司网络，请按照公司规定使用。\n请勿进行非工作下载及视频行为！",
		AccountMail: accountMail,
	}
	v, _ = json.Marshal(other)
	s = &Setting{Name: StructName(other), Data: string(v)}
	_, err = sess.InsertOne(s)
	if err != nil {
		return err
	}

	// Install
	install := &Setting{Name: InstallName, Data: InstallData}
	_, err = sess.InsertOne(install)
	if err != nil {
		return err
	}

	return sess.Commit()
}

func CheckErrNotFound(err error) bool {
	return err == ErrNotFound
}

const accountMail = `<p>您好:</p>
<p>&nbsp;&nbsp;您的{{.Issuer}}账号已经审核开通。</p>
<p>
    登陆地址: <b>{{.LinkAddr}}</b> <br/>
    用户组: <b>{{.Group}}</b> <br/>
    用户名: <b>{{.Username}}</b> <br/>
    用户PIN码: <b>{{.PinCode}}</b> <br/>
    用户动态码(3天后失效):<br/>
    <img src="{{.OtpImg}}"/>
</p>
<div>
    使用说明:
    <ul>
        <li>请使用OTP软件扫描动态码二维码</li>
        <li>然后使用anyconnect客户端进行登陆</li>
        <li>登陆密码为 【PIN码+动态码】</li>
    </ul>
</div>
<p>
    软件下载地址: https://{{.LinkAddr}}/files/info.txt
</p>`
