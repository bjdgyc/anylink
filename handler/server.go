package handler

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/bjdgyc/anylink/common"
	"github.com/bjdgyc/anylink/proxyproto"
	"github.com/bjdgyc/anylink/router"
)

func startAdmin() {
	mux := router.NewHttpMux()
	mux.HandleFunc(router.ANY, "/", notFound)
	// mux.ServeFile(router.ANY, "/static/*", http.Dir("./static"))

	// pprof
	mux.HandleFunc(router.ANY, "/debug/pprof/*", pprof.Index)
	mux.HandleFunc(router.ANY, "/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc(router.ANY, "/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc(router.ANY, "/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc(router.ANY, "/debug/pprof/trace", pprof.Trace)

	fmt.Println("Listen admin", common.ServerCfg.AdminAddr)
	err := http.ListenAndServe(common.ServerCfg.AdminAddr, mux)
	fmt.Println(err)
}

func startTls() {
	addr := common.ServerCfg.ServerAddr
	certFile := common.ServerCfg.CertFile
	keyFile := common.ServerCfg.CertKey

	// 设置tls信息
	tlsConfig := &tls.Config{
		NextProtos: []string{"http/1.1"},
		MinVersion: tls.VersionTLS12,
	}
	srv := &http.Server{
		Addr:      addr,
		Handler:   initRoute(),
		TLSConfig: tlsConfig,
	}

	var ln net.Listener

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	if common.ServerCfg.ProxyProtocol {
		ln = &proxyproto.Listener{Listener: ln, ProxyHeaderTimeout: time.Second * 5}
	}

	fmt.Println("listen ", addr)
	err = srv.ServeTLS(ln, certFile, keyFile)
	if err != nil {
		log.Fatal(err)
	}
}

func initRoute() http.Handler {
	mux := router.NewHttpMux()
	mux.HandleFunc("GET", "/", checkLinkClient(LinkHome))
	mux.HandleFunc("POST", "/", checkLinkClient(LinkAuth))
	mux.HandleFunc("CONNECT", "/CSCOSSLC/tunnel", LinkTunnel)
	mux.SetNotFound(http.HandlerFunc(notFound))
	return mux
}

func notFound(w http.ResponseWriter, r *http.Request) {
	// fmt.Println(r.RemoteAddr)
	// hu, _ := httputil.DumpRequest(r, true)
	// fmt.Println("NotFound: ", string(hu))

	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintln(w, "404 page not found")
}
