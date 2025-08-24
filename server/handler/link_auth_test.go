package handler

import (
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bjdgyc/anylink/base"
)

func TestLinkAuth_AuthCert(t *testing.T) {
	base.Test()

	//	开启证书验证但未提供证书
	base.Cfg.AuthCert = true
	base.Cfg.AuthOnlyCert = true

	req := httptest.NewRequest("POST", "/", strings.NewReader(`<?xml version="1.0" encoding="UTF-8"?><config-auth><type>auth-reply</type><auth><username>test</username><password>test</password></auth><group-select>default</group-select></config-auth>`))
	req.Header.Set("User-Agent", "cisco anyconnect vpn agent")
	req.Header.Set("X-Aggregate-Auth", "1")
	req.Header.Set("X-Transcend-Version", "1")

	w := httptest.NewRecorder()
	LinkAuth(w, req)

	if w.Code != http.StatusForbidden {
		t.Error()
	}

	// 开启证书验证但未提供证书，但证书验证失败
	base.Cfg.AuthCert = true
	base.Cfg.AuthOnlyCert = true

	cert := &x509.Certificate{
		Subject: pkix.Name{
			CommonName:         "",
			OrganizationalUnit: []string{""},
		},
	}
	req.TLS = &tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{cert},
	}

	w = httptest.NewRecorder()
	LinkAuth(w, req)

	if w.Code != http.StatusBadRequest {
		t.Error()
	}

	// 开启证书验证但未提供证书，未开启仅证书认证
	base.Cfg.AuthCert = true
	base.Cfg.AuthOnlyCert = false

	w = httptest.NewRecorder()
	LinkAuth(w, req)

	if w.Code == http.StatusForbidden {
		t.Error()
	}
}
