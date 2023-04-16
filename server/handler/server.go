package handler

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/pires/go-proxyproto"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bjdgyc/anylink/base"
	"github.com/gorilla/mux"
	"github.com/pion/dtls/v2/pkg/crypto/selfsign"
)

func startTls() {

	var (
		err error

		addr = base.Cfg.ServerAddr
		ln   net.Listener
	)

	tempCert, _ := selfsign.GenerateSelfSignedWithDNS("localhost")

	var certs []tls.Certificate
	var nameToCertificate map[string]*tls.Certificate

	// TODO 后续可以实现加载证书
	cert, _ := tls.LoadX509KeyPair(base.Cfg.CertFile, base.Cfg.CertKey)
	certs = append(certs, cert)

	nameToCertificate = buildNameToCertificate(certs)

	// 判断证书文件
	// _, err = os.Stat(certFile)
	// if errors.Is(err, os.ErrNotExist) {
	//	// 自动生成证书
	//	certs[0], err = selfsign.GenerateSelfSignedWithDNS("vpn.anylink")
	// } else {
	//	// 使用自定义证书
	//	certs[0], err = tls.LoadX509KeyPair(certFile, keyFile)
	// }

	// 修复 CVE-2016-2183
	// https://segmentfault.com/a/1190000038486901
	// nmap -sV --script ssl-enum-ciphers -p 443 www.example.com
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
		GetCertificate: func(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
			// Copy from tls.Config getCertificate()
			name := strings.ToLower(clientHello.ServerName)
			if cert, ok := nameToCertificate[name]; ok {
				return cert, nil
			}
			if len(name) > 0 {
				labels := strings.Split(name, ".")
				labels[0] = "*"
				wildcardName := strings.Join(labels, ".")
				if cert, ok := nameToCertificate[wildcardName]; ok {
					return cert, nil
				}
			}
			return &tempCert, nil
		},
		// InsecureSkipVerify: true,
	}
	srv := &http.Server{
		Addr:      addr,
		Handler:   initRoute(),
		TLSConfig: tlsConfig,
		ErrorLog:  base.GetBaseLog(),
	}

	ln, err = net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	if base.Cfg.ProxyProtocol {
		ln = &proxyproto.Listener{
			Listener:          ln,
			ReadHeaderTimeout: 40 * time.Second,
		}
	}

	base.Info("listen server", addr)
	err = srv.ServeTLS(ln, "", "")
	if err != nil {
		base.Fatal(err)
	}
}

// Copy from tls.Config BuildNameToCertificate()
func buildNameToCertificate(certificates []tls.Certificate) map[string]*tls.Certificate {
	var certMap = make(map[string]*tls.Certificate)
	for i := range certificates {
		cert := &certificates[i]
		x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
		if err != nil {
			continue
		}
		startTime := x509Cert.NotBefore.String()
		expiredTime := x509Cert.NotAfter.String()
		if x509Cert.Subject.CommonName != "" && len(x509Cert.DNSNames) == 0 {
			commonName := x509Cert.Subject.CommonName
			fmt.Printf("┏ Load Certificate: %s\n", commonName)
			fmt.Printf("┠╌╌ Start Time:     %s\n", startTime)
			fmt.Printf("┖╌╌ Expired Time:   %s\n", expiredTime)
			certMap[commonName] = cert
		}
		for _, san := range x509Cert.DNSNames {
			fmt.Printf("┏ Load Certificate: %s\n", san)
			fmt.Printf("┠╌╌ Start Time:     %s\n", startTime)
			fmt.Printf("┖╌╌ Expired Time:   %s\n", expiredTime)
			certMap[san] = cert
		}
	}
	return certMap
}

func initRoute() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/", LinkHome).Methods(http.MethodGet)
	r.HandleFunc("/", LinkAuth).Methods(http.MethodPost)
	r.HandleFunc("/CSCOSSLC/tunnel", LinkTunnel).Methods(http.MethodConnect)
	r.HandleFunc("/otp_qr", LinkOtpQr).Methods(http.MethodGet)
	r.HandleFunc("/profile.xml", func(w http.ResponseWriter, r *http.Request) {
		b, _ := os.ReadFile(base.Cfg.Profile)
		w.Write(b)
	}).Methods(http.MethodGet)
	r.PathPrefix("/files/").Handler(
		http.StripPrefix("/files/",
			http.FileServer(http.Dir(base.Cfg.FilesPath)),
		),
	)
	// 健康检测
	r.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "ok")
	}).Methods(http.MethodGet)
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
