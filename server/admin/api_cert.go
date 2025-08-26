package admin

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/bjdgyc/anylink/base"
	"github.com/bjdgyc/anylink/dbdata"
)

func CustomCert(w http.ResponseWriter, r *http.Request) {
	cert, _, err := r.FormFile("cert")
	if err != nil {
		RespError(w, RespInternalErr, err)
		return
	}
	key, _, err := r.FormFile("key")
	if err != nil {
		RespError(w, RespInternalErr, err)
		return
	}
	certFile, err := os.OpenFile(base.Cfg.CertFile, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0600)
	if err != nil {
		RespError(w, RespInternalErr, err)
		return
	}
	defer certFile.Close()
	if _, err := io.Copy(certFile, cert); err != nil {
		RespError(w, RespInternalErr, err)
		return
	}
	keyFile, err := os.OpenFile(base.Cfg.CertKey, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0600)
	if err != nil {
		RespError(w, RespInternalErr, err)
		return
	}
	defer keyFile.Close()
	if _, err := io.Copy(keyFile, key); err != nil {
		RespError(w, RespInternalErr, err)
		return
	}
	if tlscert, _, err := dbdata.ParseCert(); err != nil {
		RespError(w, RespInternalErr, fmt.Sprintf("证书不合法，请重新上传:%v", err))
		return
	} else {
		dbdata.LoadCertificate(tlscert)
	}
	RespSucess(w, "上传成功")
}
func GetCertSetting(w http.ResponseWriter, r *http.Request) {
	sess := dbdata.GetXdb().NewSession()
	defer sess.Close()
	data := &dbdata.SettingLetsEncrypt{}
	if err := dbdata.SettingGet(data); err != nil {
		dbdata.SettingSessAdd(sess, data)
		RespError(w, RespInternalErr, err)
	}
	userData := &dbdata.LegoUserData{}
	if err := dbdata.SettingGet(userData); err != nil {
		dbdata.SettingSessAdd(sess, userData)
	}
	RespSucess(w, data)
}
func CreatCert(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		RespError(w, RespInternalErr, err)
		return
	}
	defer r.Body.Close()
	config := &dbdata.SettingLetsEncrypt{}
	if err := json.Unmarshal(body, config); err != nil {
		RespError(w, RespInternalErr, err)
		return
	}
	if err := dbdata.SettingSet(config); err != nil {
		RespError(w, RespInternalErr, err)
		return
	}
	client := dbdata.LeGoClient{}
	if err := client.NewClient(config); err != nil {
		base.Error(err)
		RespError(w, RespInternalErr, fmt.Sprintf("获取证书失败:%v", err))
		return
	}
	if err := client.GetCert(config.Domain); err != nil {
		base.Error(err)
		RespError(w, RespInternalErr, fmt.Sprintf("获取证书失败:%v", err))
		return
	}
	RespSucess(w, "生成证书成功")
}

// 初始化客户端 CA
func InitClientCA(w http.ResponseWriter, r *http.Request) {
	// 检查 CA 文件是否已存在
	caExists := true
	if _, err := os.Stat(base.Cfg.ClientCertCAFile); errors.Is(err, os.ErrNotExist) {
		caExists = false
	}
	keyExists := true
	if _, err := os.Stat(base.Cfg.ClientCertCAKeyFile); errors.Is(err, os.ErrNotExist) {
		keyExists = false
	}

	if caExists && keyExists {
		RespError(w, RespInternalErr, "客户端 CA 已存在，请勿重复初始化,如需强制初始化可在服务器后台删除客户端CA文件")
		return
	}
	err := dbdata.GenerateClientCA()
	if err != nil {
		RespError(w, RespInternalErr, fmt.Sprintf("客户端 CA 生成失败: %v", err))
		return
	}
	RespSucess(w, "客户端 CA 初始化成功")
}

// 生成客户端证书
func GenerateClientCert(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	if username == "" {
		RespError(w, RespInternalErr, "用户名不能为空")
		return
	}
	groupname := r.FormValue("group_name")
	if groupname == "" {
		RespError(w, RespInternalErr, "用户组不能为空")
		return
	}

	// 检查用户是否存在
	user := &dbdata.User{}
	err := dbdata.One("Username", username, user)
	if err != nil {
		RespError(w, RespInternalErr, "用户不存在")
		return
	}

	// 生成客户端证书
	certData, err := dbdata.GenerateClientCert(username, groupname)
	if err != nil {
		RespError(w, RespInternalErr, fmt.Sprintf("证书生成失败: %v", err))
		return
	}

	RespSucess(w, certData)
}

// 下载客户端 P12 证书
func DownloadClientP12(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	groupname := r.FormValue("groupname")
	password := r.FormValue("password")

	if username == "" {
		RespError(w, RespInternalErr, "用户名不能为空")
		return
	}
	if groupname == "" {
		RespError(w, RespInternalErr, "用户组不能为空")
		return
	}

	// if password == "" {
	// 	password = "123456" // 默认密码
	// }

	// 生成 P12 证书
	p12Data, err := dbdata.GenerateClientP12FromDB(username, groupname, password)
	if err != nil {
		RespError(w, RespInternalErr, fmt.Sprintf("证书下载失败: %v", err))
		return
	}

	// 设置下载响应头
	w.Header().Set("Content-Type", "application/x-pkcs12")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.p12", username))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(p12Data)))
	w.Write(p12Data)
}

// 切换客户端证书状态（禁用/启用）
func ChangeClientCertStatus(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	if username == "" {
		RespError(w, RespInternalErr, "用户名不能为空")
		return
	}
	groupname := r.FormValue("groupname")
	if groupname == "" {
		RespError(w, RespInternalErr, "用户组不能为空")
		return
	}

	clientCert, err := dbdata.GetClientCert(username, groupname)
	if err != nil {
		RespError(w, RespInternalErr, "证书不存在")
		return
	}

	err = clientCert.ChangeStatus()
	if err != nil {
		RespError(w, RespInternalErr, fmt.Sprintf("证书状态切换失败: %v", err))
		return
	}

	statusText := "启用"
	if clientCert.Status == dbdata.CertStatusDisabled {
		statusText = "禁用"
	}

	RespSucess(w, fmt.Sprintf("证书%s成功", statusText))
}

// 删除客户端证书
func DeleteClientCert(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	if username == "" {
		RespError(w, RespInternalErr, "用户名不能为空")
		return
	}
	groupname := r.FormValue("groupname")
	if groupname == "" {
		RespError(w, RespInternalErr, "用户组不能为空")
		return
	}

	clientCert, err := dbdata.GetClientCert(username, groupname)
	if err != nil {
		RespError(w, RespInternalErr, "证书不存在")
		return
	}

	err = clientCert.Delete()
	if err != nil {
		RespError(w, RespInternalErr, fmt.Sprintf("证书删除失败: %v", err))
		return
	}

	RespSucess(w, "证书删除成功")
}

// 获取客户端证书列表
func GetClientCertList(w http.ResponseWriter, r *http.Request) {
	pageSize := 10
	pageIndex := 1

	if r.FormValue("page_size") != "" {
		if ps, err := strconv.Atoi(r.FormValue("page_size")); err == nil {
			pageSize = ps
		}
	}

	if r.FormValue("page_index") != "" {
		if pi, err := strconv.Atoi(r.FormValue("page_index")); err == nil {
			pageIndex = pi
		}
	}

	// 添加搜索参数
	username := r.FormValue("username")
	groupname := r.FormValue("groupname")
	status := r.FormValue("status")

	certs, total, err := dbdata.GetClientCertList(pageSize, pageIndex, username, groupname, status)
	if err != nil {
		RespError(w, RespInternalErr, fmt.Sprintf("获取证书列表失败: %v", err))
		return
	}

	data := map[string]any{
		"list":  certs,
		"total": total,
	}

	RespSucess(w, data)
}

// UserCertInfo 获取用户证书生成所需信息
func UserCertInfo(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()

	// 获取所有启用的用户
	var users []dbdata.User
	err := dbdata.Find(&users, 1000, 1)
	if err != nil && !dbdata.CheckErrNotFound(err) {
		RespError(w, RespInternalErr, err)
		return
	}

	// 获取所有启用的组
	var groups []dbdata.Group
	err = dbdata.Find(&groups, 1000, 1)
	if err != nil && !dbdata.CheckErrNotFound(err) {
		RespError(w, RespInternalErr, err)
		return
	}

	// 过滤启用的用户和组
	activeUsers := make([]dbdata.User, 0)
	for _, user := range users {
		if user.Status == 1 {
			activeUsers = append(activeUsers, user)
		}
	}

	activeGroups := make([]dbdata.Group, 0)
	for _, group := range groups {
		if group.Status == 1 {
			activeGroups = append(activeGroups, group)
		}
	}

	data := map[string]any{
		"users":  activeUsers,
		"groups": activeGroups,
	}

	RespSucess(w, data)
}
