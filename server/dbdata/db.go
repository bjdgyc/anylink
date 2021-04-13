package dbdata

import (
	"time"

	"github.com/asdine/storm/v3"
	"github.com/asdine/storm/v3/codec/json"
	"github.com/bjdgyc/anylink/base"
	bolt "go.etcd.io/bbolt"
)

var (
	sdb *storm.DB
)

func initDb() {
	var err error
	sdb, err = storm.Open(base.Cfg.DbFile, storm.Codec(json.Codec),
		storm.BoltOptions(0600, &bolt.Options{Timeout: 10 * time.Second}))
	if err != nil {
		base.Fatal(err)
	}

	// 初始化数据库
	err = sdb.Init(&User{})
	if err != nil {
		base.Fatal(err)
	}

	// fmt.Println("s1")
}

func initData() {
	var (
		err     error
		install bool
	)

	// 判断是否初次使用
	err = Get(SettingBucket, Installed, &install)
	if err == nil && install {
		// 已经安装过
		return
	}

	defer func() {
		_ = Set(SettingBucket, Installed, true)
	}()

	smtp := &SettingSmtp{
		Host: "127.0.0.1",
		Port: 25,
		From: "vpn@xx.com",
	}
	_ = SettingSet(smtp)

	other := &SettingOther{
		LinkAddr:    "vpn.xx.com",
		Banner:      "您已接入公司网络，请按照公司规定使用。\n请勿进行非工作下载及视频行为！",
		AccountMail: accountMail,
	}
	_ = SettingSet(other)

}

func CheckErrNotFound(err error) bool {
	return err == storm.ErrNotFound
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
