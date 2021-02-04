package admin

import (
	"fmt"
	"net/http"
	"time"

	"github.com/bjdgyc/anylink/base"
	"github.com/bjdgyc/anylink/pkg/utils"
	"github.com/gorilla/mux"
)

// 登陆接口
func Login(w http.ResponseWriter, r *http.Request) {
	// TODO 调试信息输出
	// hd, _ := httputil.DumpRequest(r, true)
	// fmt.Println("DumpRequest: ", string(hd))

	_ = r.ParseForm()
	admin_user := r.PostFormValue("admin_user")
	admin_pass := r.PostFormValue("admin_pass")

	// 认证错误
	if !(admin_user == base.Cfg.AdminUser &&
		utils.PasswordVerify(admin_pass, base.Cfg.AdminPass)) {
		RespError(w, RespUserOrPassErr)
		return
	}

	// token有效期
	expiresAt := time.Now().Unix() + 3600*3
	jwtData := map[string]interface{}{"admin_user": admin_user}
	tokenString, err := SetJwtData(jwtData, expiresAt)
	if err != nil {
		RespError(w, 1, err)
		return
	}

	data := make(map[string]interface{})
	data["token"] = tokenString
	data["admin_user"] = admin_user
	data["expires_at"] = expiresAt

	RespSucess(w, data)
}

func authMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == http.MethodOptions {
			return
		}

		route := mux.CurrentRoute(r)
		name := route.GetName()
		// fmt.Println("bb", r.URL.Path, name)
		if utils.InArrStr([]string{"login", "index", "static"}, name) {
			// 不进行鉴权
			next.ServeHTTP(w, r)
			return
		}

		// 进行登陆鉴权
		jwtToken := r.Header.Get("Jwt")
		if jwtToken == "" {
			jwtToken = r.FormValue("jwt")
		}
		data, err := GetJwtData(jwtToken)
		if err != nil || base.Cfg.AdminUser != fmt.Sprint(data["admin_user"]) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
