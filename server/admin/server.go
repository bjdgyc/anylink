// admin:后台管理接口
package admin

import (
	"embed"
	"net/http"
	"net/http/pprof"

	"github.com/bjdgyc/anylink/base"
	"github.com/gorilla/mux"
)

var UiData embed.FS

// StartAdmin 开启服务
func StartAdmin() {

	r := mux.NewRouter()
	r.Use(authMiddleware)

	// 监控检测
	r.HandleFunc("/status.html", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}).Name("index")

	r.Handle("/", http.RedirectHandler("/ui/", http.StatusFound)).Name("index")
	r.PathPrefix("/ui/").Handler(
		// http.StripPrefix("/ui/", http.FileServer(http.Dir(base.Cfg.UiPath))),
		http.FileServer(http.FS(UiData)),
	).Name("static")
	r.HandleFunc("/base/login", Login).Name("login")

	r.HandleFunc("/set/home", SetHome)
	r.HandleFunc("/set/system", SetSystem)
	r.HandleFunc("/set/soft", SetSoft)
	r.HandleFunc("/set/other", SetOther)
	r.HandleFunc("/set/other/edit", SetOtherEdit)
	r.HandleFunc("/set/other/smtp", SetOtherSmtp)
	r.HandleFunc("/set/other/smtp/edit", SetOtherSmtpEdit)
	r.HandleFunc("/set/audit/list", SetAuditList)

	r.HandleFunc("/user/list", UserList)
	r.HandleFunc("/user/detail", UserDetail)
	r.HandleFunc("/user/set", UserSet)
	r.HandleFunc("/user/del", UserDel)
	r.HandleFunc("/user/online", UserOnline)
	r.HandleFunc("/user/offline", UserOffline)
	r.HandleFunc("/user/reline", UserReline)
	r.HandleFunc("/user/otp_qr", UserOtpQr)
	r.HandleFunc("/user/ip_map/list", UserIpMapList)
	r.HandleFunc("/user/ip_map/detail", UserIpMapDetail)
	r.HandleFunc("/user/ip_map/set", UserIpMapSet)
	r.HandleFunc("/user/ip_map/del", UserIpMapDel)

	r.HandleFunc("/group/list", GroupList)
	r.HandleFunc("/group/names", GroupNames)
	r.HandleFunc("/group/detail", GroupDetail)
	r.HandleFunc("/group/set", GroupSet)
	r.HandleFunc("/group/del", GroupDel)

	// pprof
	if base.Cfg.Pprof {
		r.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline).Name("debug")
		r.HandleFunc("/debug/pprof/profile", pprof.Profile).Name("debug")
		r.HandleFunc("/debug/pprof/symbol", pprof.Symbol).Name("debug")
		r.HandleFunc("/debug/pprof/trace", pprof.Trace).Name("debug")
		r.HandleFunc("/debug/pprof", location("/debug/pprof/")).Name("debug")
		r.PathPrefix("/debug/pprof/").HandlerFunc(pprof.Index).Name("debug")
	}

	base.Info("Listen admin", base.Cfg.AdminAddr)
	err := http.ListenAndServeTLS(base.Cfg.AdminAddr, base.Cfg.CertFile, base.Cfg.CertKey, r)
	if err != nil {
		base.Fatal(err)
	}
}

func location(url string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", url)
		w.WriteHeader(http.StatusFound)
	}
}
