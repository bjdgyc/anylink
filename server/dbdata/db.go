package dbdata

import (
	"net/http"
	"time"

	"github.com/bjdgyc/anylink/base"
	_ "github.com/denisenkom/go-mssqldb"
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
	if err != nil {
		base.Fatal(err)
	}

	// 初始化xorm时区
	xdb.DatabaseTZ = time.Local
	xdb.TZLocation = time.Local

	if base.Cfg.ShowSQL {
		xdb.ShowSQL(true)
	}

	// 初始化数据库
	err = xdb.Sync2(&User{}, &Setting{}, &Group{}, &IpMap{}, &AccessAudit{}, &Policy{}, &StatsNetwork{}, &StatsCpu{}, &StatsMem{}, &StatsOnline{}, &UserActLog{}, &PasswordReset{})
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
	install := &SettingInstall{}
	err = SettingGet(install)

	if err == nil && install.Installed {
		// 已经安装过
		return
	}

	// 发生错误
	if err != ErrNotFound {
		base.Fatal(err)
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
	err = SettingSessAdd(sess, smtp)
	if err != nil {
		return err
	}

	// SettingAuditLog
	auditLog := SettingGetAuditLogDefault()
	err = SettingSessAdd(sess, auditLog)
	if err != nil {
		return err
	}

	// SettingDnsProvider
	provider := &SettingLetsEncrypt{
		Domain:      "vpn.xxx.com",
		Legomail:    "legomail",
		Name:        "aliyun",
		Renew:       false,
		DNSProvider: DNSProvider{},
	}
	err = SettingSessAdd(sess, provider)
	if err != nil {
		return err
	}
	// LegoUser
	legouser := &LegoUserData{}
	err = SettingSessAdd(sess, legouser)
	if err != nil {
		return err
	}
	// SettingOther
	other := &SettingOther{
		LinkAddr:    "vpn.xx.com",
		Banner:      "您已接入公司网络，请按照公司规定使用。\n请勿进行非工作下载及视频行为！",
		Homecode:    http.StatusOK,
		Homeindex:   "AnyLink 是一个企业级远程办公 sslvpn 的软件，可以支持多人同时在线使用。",
		AccountMail: accountMail,
	}
	err = SettingSessAdd(sess, other)
	if err != nil {
		return err
	}

	// Install
	install := &SettingInstall{Installed: true}
	err = SettingSessAdd(sess, install)
	if err != nil {
		return err
	}

	err = sess.Commit()
	if err != nil {
		return err
	}

	g1 := Group{
		Name:         "all",
		AllowLan:     true,
		ClientDns:    []ValData{{Val: "114.114.114.114"}},
		RouteInclude: []ValData{{Val: ALL}},
		Status:       1,
	}
	err = SetGroup(&g1)
	if err != nil {
		return err
	}

	g2 := Group{
		Name:         "ops",
		AllowLan:     true,
		ClientDns:    []ValData{{Val: "114.114.114.114"}},
		RouteInclude: []ValData{{Val: "10.0.0.0/8"}},
		Status:       1,
	}
	err = SetGroup(&g2)
	if err != nil {
		return err
	}

	return nil
}

func CheckErrNotFound(err error) bool {
	return err == ErrNotFound
}

// base64 图片
// 用户动态码(请妥善保存):<br/>
// <img src="{{.OtpImgBase64}}"/><br/>
const accountMail = `<p>您好:</p>
<p>&nbsp;&nbsp;您的{{.Issuer}}账号已经审核开通。</p>
<p>
    登陆地址: <b>{{.LinkAddr}}</b> <br/>
    用户组: <b>{{.Group}}</b> <br/>
    用户名: <b>{{.Username}}</b> <br/>
    用户PIN码: <b>{{.PinCode}}</b> <br/>
    用户过期时间: <b>{{.LimitTime}}</b> <br/>
    {{if .DisableOtp}}
    <!-- nothing -->
    {{else}}
	
    <!-- 
    用户动态码(3天后失效):<br/>
    <img src="{{.OtpImg}}"/><br/>
    -->
    用户动态码(请妥善保存):<br/>
    <img src="cid:userOtpQr.png" alt="userOtpQr" /><br/>

    {{end}}
</p>
<div>
    使用说明:
    <ul>
        <li>请使用OTP软件扫描动态码二维码</li>
        <li>然后使用anyconnect客户端进行登陆</li>
        <li>登陆密码为 PIN 码</li>
		<li>OTP密码为扫码后生成的动态码</li>
    </ul>
</div>
<p>
    软件下载地址: https://{{.LinkAddr}}/files/info.txt
</p>`
