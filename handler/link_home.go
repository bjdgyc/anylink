package handler

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"strings"
)

func LinkHome(w http.ResponseWriter, r *http.Request) {
	hu, _ := httputil.DumpRequest(r, true)
	fmt.Println("DumpHome: ", string(hu))
	fmt.Println(r.RemoteAddr)

	connection := strings.ToLower(r.Header.Get("Connection"))
	userAgent := strings.ToLower(r.UserAgent())
	if connection == "close" && strings.Contains(userAgent, "anyconnect") {
		w.Header().Set("Connection", "close")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "hello world")
}
