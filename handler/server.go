package handler

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	_ "net/http/pprof"

	"github.com/bjdgyc/anylink/common"
	"github.com/julienschmidt/httprouter"
)

func Start() {
	testTun()
	go startDebug()
	go startDtls()
	go startTls()
}

func startDebug() {
	http.ListenAndServe(common.ServerCfg.DebugAddr, nil)
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

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	srv.SetKeepAlivesEnabled(true)
	fmt.Println("listen ", addr)
	err = srv.ServeTLS(ln, certFile, keyFile)
	if err != nil {
		log.Fatal(err)
	}
}

func initRoute() http.Handler {
	router := httprouter.New()
	router.GET("/", checkVpnClient(LinkHome))
	router.POST("/", checkVpnClient(LinkAuth))
	router.HandlerFunc("CONNECT", "/CSCOSSLC/tunnel", LinkTunnel)
	router.NotFound = http.HandlerFunc(notFound)
	return router
}

func notFound(w http.ResponseWriter, r *http.Request) {
	hu, _ := httputil.DumpRequest(r, true)
	fmt.Println("NotFound: ", string(hu))

	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintln(w, "404 page not found")
}
