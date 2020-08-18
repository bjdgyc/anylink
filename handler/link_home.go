package handler

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/julienschmidt/httprouter"
)

func LinkHome(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	hu, _ := httputil.DumpRequest(r, true)
	fmt.Println("DumpHome: ", string(hu))
	fmt.Println(r.RemoteAddr)

	connection := strings.ToLower(r.Header.Get("Connection"))
	if connection == "close" {
		w.Header().Set("Connection", "close")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "hello world")
}
