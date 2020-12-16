package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/bjdgyc/anylink/admin"
)

func LinkHome(w http.ResponseWriter, r *http.Request) {
	// fmt.Println(r.RemoteAddr)
	// hu, _ := httputil.DumpRequest(r, true)
	// fmt.Println("DumpHome: ", string(hu))

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

func LinkOtpQr(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	idS := r.FormValue("id")
	jwtToken := r.FormValue("jwt")
	data, err := admin.GetJwtData(jwtToken)
	if err != nil || idS != fmt.Sprint(data["id"]) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	admin.UserOtpQr(w, r)
}
