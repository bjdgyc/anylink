package dbdata

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bjdgyc/anylink/base"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

//定义orm引擎
var x *xorm.Engine

//创建orm引擎
func initDb() {
	var err error
	dbconfig := base.Cfg.DbFile

	configlist := strings.Split(dbconfig, ":")
	if len(configlist) < 2 {
		log.Println("数据库配置错误 :", configlist, "，自动 使用默认sqlite3数据库")
		x, err = xorm.NewEngine("sqlite3", "./test.db")

	}
	x, err = xorm.NewEngine(configlist[0], strings.Join(configlist[1:], ":"))

	//x, err = xorm.NewEngine("mysql", "root:zg1234567899@tcp(172.16.249.34:3306)/test?charset=utf8")
	//x, err = xorm.NewEngine("sqlite3", "./test.db")
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}
	if err := x.Sync2(new(Group), new(User), new(SettingSmtp), new(SettingOther), new(IpMap)); err != nil {
		log.Fatal("数据表同步失败:", err)
	}
	x.SetConnMaxLifetime(time.Hour)
	//x.ShowSQL(true)
	x.ShowExecTime(true)
	x.SetMaxIdleConns(10)
	x.SetMaxOpenConns(50)
	// cacher := xorm.NewLRUCacher(xorm.NewMemoryStore(), 10000)
	// x.SetDefaultCacher(cacher)
}

func initData() {
	// 判断是否初次使用
	n1 := CountAll(&SettingSmtp{})
	fmt.Println(n1)
	if n1 > 0 {
		// 已经安装过
		return
	}
	smtp := &SettingSmtp{
		Host: "127.0.0.1",
		Port: 25,
		From: "vpn@xx.com",
	}
	err := Save(smtp)
	fmt.Println(err)

	other := &SettingOther{
		LinkAddr:    "vpn.xx.com",
		Banner:      "您已接入公司网络，请按照公司规定使用。\n请勿进行非工作下载及视频行为！",
		AccountMail: accountMail,
	}
	err = Save(other)
	fmt.Println(err)
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
