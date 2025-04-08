// admin:后台管理接口
package admin

import (
	"crypto/tls"
	"embed"
	"net/http"
	"net/http/pprof"

	"github.com/arl/statsviz"
	"github.com/bjdgyc/anylink/base"
	"github.com/bjdgyc/anylink/dbdata"
	"github.com/bjdgyc/anylink/pkg/utils"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

var UiData embed.FS

// StartAdmin 开启服务
func StartAdmin() {

	r := mux.NewRouter()
	r.Use(recoverHttp, authMiddleware, handlers.CompressHandler)
	// 所有路由添加安全头
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			utils.SetSecureHeader(w)
			w.Header().Set("Server", "AnyLinkAdminOpenSource")
			next.ServeHTTP(w, req)
		})
	})

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
	r.HandleFunc("/set/other/audit_log", SetOtherAuditLog)
	r.HandleFunc("/set/other/audit_log/edit", SetOtherAuditLogEdit)
	r.HandleFunc("/set/audit/list", SetAuditList)
	r.HandleFunc("/set/audit/export", SetAuditExport)
	r.HandleFunc("/set/audit/act_log_list", UserActLogList)
	r.HandleFunc("/set/other/createcert", CreatCert)
	r.HandleFunc("/set/other/getcertset", GetCertSetting)
	r.HandleFunc("/set/other/customcert", CustomCert)

	r.HandleFunc("/user/list", UserList)
	r.HandleFunc("/user/detail", UserDetail)
	r.HandleFunc("/user/set", UserSet)
	r.HandleFunc("/user/uploaduser", UserUpload).Methods(http.MethodPost)
	r.HandleFunc("/user/del", UserDel)
	r.HandleFunc("/user/online", UserOnline)
	r.HandleFunc("/user/offline", UserOffline)
	r.HandleFunc("/user/reline", UserReline)
	r.HandleFunc("/user/otp_qr", UserOtpQr)
	r.HandleFunc("/user/ip_map/list", UserIpMapList)
	r.HandleFunc("/user/ip_map/detail", UserIpMapDetail)
	r.HandleFunc("/user/ip_map/set", UserIpMapSet)
	r.HandleFunc("/user/ip_map/del", UserIpMapDel)
	r.HandleFunc("/user/policy/list", PolicyList)
	r.HandleFunc("/user/policy/detail", PolicyDetail)
	r.HandleFunc("/user/policy/set", PolicySet)
	r.HandleFunc("/user/policy/del", PolicyDel)
	r.HandleFunc("/user/reset/forgotPassword", ForgotPassword).Name("forgot_password")
	r.HandleFunc("/user/reset/resetPassword", ResetPassword).Name("reset_password")

	r.HandleFunc("/group/list", GroupList)
	r.HandleFunc("/group/names", GroupNames)
	r.HandleFunc("/group/names_ids", GroupNamesIds)
	r.HandleFunc("/group/detail", GroupDetail)
	r.HandleFunc("/group/set", GroupSet)
	r.HandleFunc("/group/del", GroupDel)
	r.HandleFunc("/group/auth_login", GroupAuthLogin)

	r.HandleFunc("/statsinfo/list", StatsInfoList)
	r.HandleFunc("/locksinfo/list", GetLocksInfo)
	r.HandleFunc("/locksinfo/unlok", UnlockUser)

	// pprof
	if base.Cfg.Pprof {
		r.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline).Name("debug")
		r.HandleFunc("/debug/pprof/profile", pprof.Profile).Name("debug")
		r.HandleFunc("/debug/pprof/symbol", pprof.Symbol).Name("debug")
		r.HandleFunc("/debug/pprof/trace", pprof.Trace).Name("debug")
		r.HandleFunc("/debug/pprof", location("/debug/pprof/")).Name("debug")
		r.PathPrefix("/debug/pprof/").HandlerFunc(pprof.Index).Name("debug")
		// statsviz
		srv, _ := statsviz.NewServer() // Create server or handle error
		r.Path("/debug/statsviz/ws").Name("debug").HandlerFunc(srv.Ws())
		r.PathPrefix("/debug/statsviz/").Name("debug").Handler(srv.Index())
	}

	base.Info("Listen admin", base.Cfg.AdminAddr)

	// 修复 CVE-2016-2183
	cipherSuites := tls.CipherSuites()
	selectedCipherSuites := make([]uint16, 0, len(cipherSuites))
	for _, s := range cipherSuites {
		selectedCipherSuites = append(selectedCipherSuites, s.ID)
	}

	// 设置tls信息
	tlsConfig := &tls.Config{
		NextProtos:   []string{"http/1.1"},
		MinVersion:   tls.VersionTLS12,
		CipherSuites: selectedCipherSuites,
		GetCertificate: func(chi *tls.ClientHelloInfo) (*tls.Certificate, error) {
			return dbdata.GetCertificateBySNI(chi.ServerName)
		},
	}
	srv := &http.Server{
		Addr:      base.Cfg.AdminAddr,
		Handler:   r,
		TLSConfig: tlsConfig,
		ErrorLog:  base.GetServerLog(),
	}
	err := srv.ListenAndServeTLS("", "")
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
