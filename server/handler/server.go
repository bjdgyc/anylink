package handler

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/bjdgyc/anylink/base"
	"github.com/bjdgyc/anylink/pkg/proxyproto"
	"github.com/gorilla/mux"
)

func GetCertificate(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	cert, err := tls.LoadX509KeyPair(base.Cfg.CertFile, base.Cfg.CertKey)
	return &cert, err
}

func startTls() {
	addr := base.Cfg.ServerAddr
	certFile := base.Cfg.CertFile
	keyFile := base.Cfg.CertKey

	// 设置tls信息
	tlsConfig := &tls.Config{
		NextProtos:         []string{"http/1.1"},
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: true,
		GetCertificate:     GetCertificate,
	}
	srv := &http.Server{
		Addr:      addr,
		Handler:   initRoute(),
		TLSConfig: tlsConfig,
		ErrorLog:  base.GetBaseLog(),
	}

	var ln net.Listener

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	if base.Cfg.ProxyProtocol {
		ln = &proxyproto.Listener{Listener: ln, ProxyHeaderTimeout: time.Second * 5}
	}

	base.Info("listen server", addr)
	err = srv.ServeTLS(ln, certFile, keyFile)
	if err != nil {
		base.Fatal(err)
	}
}

func initRoute() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/", LinkHome).Methods(http.MethodGet)
	r.HandleFunc("/", LinkAuth).Methods(http.MethodPost)
	r.HandleFunc("/CSCOSSLC/tunnel", LinkTunnel).Methods(http.MethodConnect)
	r.HandleFunc("/otp_qr", LinkOtpQr).Methods(http.MethodGet)
	r.PathPrefix("/files/").Handler(
		http.StripPrefix("/files/",
			http.FileServer(http.Dir(base.Cfg.FilesPath)),
		),
	)
	r.NotFoundHandler = http.HandlerFunc(notFound)
	return r
}

func notFound(w http.ResponseWriter, r *http.Request) {
	// fmt.Println(r.RemoteAddr)
	// hu, _ := httputil.DumpRequest(r, true)
	// fmt.Println("NotFound: ", string(hu))

	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintln(w, "404 page not found")
}
