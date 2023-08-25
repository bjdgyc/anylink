package utils

import "net/http"

// 设置安全的header头
func SetSecureHeader(w http.ResponseWriter) {
	// Content-Length Date 默认已经存在
	w.Header().Set("Server", "AnyLinkOpenSource")
	// w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("X-Aggregate-Auth", "1")

	w.Header().Set("Cache-Control", "no-store,no-cache")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Frame-Options", "deny")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Download-Options", "noopen")
	w.Header().Set("Content-Security-Policy", "default-src 'self' 'unsafe-inline'")
	w.Header().Set("X-Permitted-Cross-Domain-Policies", "none")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("Cross-Origin-Embedder-Policy", "require-corp")
	w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
	w.Header().Set("Cross-Origin-Resource-Policy", "same-origin")
	w.Header().Set("X-XSS-Protection", "0")
	w.Header().Set("Strict-Transport-Security", "max-age=31536000")

	// w.Header().Set("Clear-Site-Data", "cache,cookies,storage")
}
