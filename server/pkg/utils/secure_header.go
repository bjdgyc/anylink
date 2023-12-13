package utils

import "net/http"

// SetSecureHeader 设置安全的header头
// https://blog.csdn.net/liwan09/article/details/130248003
// https://zhuanlan.zhihu.com/p/335165168
func SetSecureHeader(w http.ResponseWriter) {
	// Content-Length Date 默认已经存在
	w.Header().Set("Server", "AnyLinkOpenSource")
	// w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("X-Aggregate-Auth", "1")

	w.Header().Set("Cache-Control", "no-store,no-cache")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Frame-Options", "SAMEORIGIN")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Download-Options", "noopen")
	w.Header().Set("Content-Security-Policy", "default-src 'self' 'unsafe-inline' 'unsafe-eval' data: blob:; frame-ancestors 'self'; base-uri 'self'; block-all-mixed-content")
	w.Header().Set("X-Permitted-Cross-Domain-Policies", "none")
	w.Header().Set("Referrer-Policy", "same-origin")
	w.Header().Set("Cross-Origin-Embedder-Policy", "require-corp")
	w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
	w.Header().Set("Cross-Origin-Resource-Policy", "same-origin")
	w.Header().Set("X-XSS-Protection", "1;mode=block")
	w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

	// w.Header().Set("Clear-Site-Data", "cache,cookies,storage")

}
