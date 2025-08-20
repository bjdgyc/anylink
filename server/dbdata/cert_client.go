package dbdata

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"os"
	"slices"
	"sync"
	"time"

	"github.com/bjdgyc/anylink/base"
	"software.sslmate.com/src/go-pkcs12"
)

// 客户端证书数据结构
type ClientCertData struct {
	Id           int       `json:"id" xorm:"pk autoincr not null"`
	Username     string    `json:"username" xorm:"varchar(60) not null"`
	GroupName    string    `json:"groupname" xorm:"varchar(60)"`
	Status       int       `json:"status" xorm:"int default 0"`
	Certificate  string    `json:"certificate" xorm:"text not null"`
	PrivateKey   string    `json:"private_key" xorm:"text not null"`
	SerialNumber string    `json:"serial_number" xorm:"varchar(100) not null"`
	NotAfter     time.Time `json:"not_after" xorm:"datetime not null"`
	CreatedAt    time.Time `json:"created_at" xorm:"datetime created"`
}

var (
	clientCACert *x509.Certificate
	clientCAKey  *rsa.PrivateKey
	caMutex      sync.Mutex
)

// 证书状态
const (
	CertStatusActive   = 0 // 有效
	CertStatusDisabled = 1 // 禁用
	CertStatusExpired  = 2 // 过期
)

// 获取证书状态描述
func (c *ClientCertData) GetStatusText() string {
	switch c.GetStatus() {
	case CertStatusActive:
		return "有效"
	case CertStatusDisabled:
		return "禁用"
	case CertStatusExpired:
		return "过期"
	default:
		return "未知"
	}
}

// 获取证书状态
func (c *ClientCertData) GetStatus() int {
	return c.Status
}

// 保存客户端证书
func (c *ClientCertData) Save() error {
	return Add(c)
}

// 禁用证书
func (c *ClientCertData) Disable() error {
	return c.UpdateStatus(CertStatusDisabled)
}

// 启用证书
func (c *ClientCertData) Enable() error {
	return c.UpdateStatus(CertStatusActive)
}

// 删除证书记录
func (c *ClientCertData) Delete() error {
	return Del(c)
}

// 切换证书状态
func (c *ClientCertData) ChangeStatus() error {
	switch c.Status {
	case CertStatusActive:
		return c.Disable()
	case CertStatusDisabled:
		return c.Enable()
	}
	return fmt.Errorf("证书已过期，无法切换状态")
}

// 更新客户端证书状态
func (c *ClientCertData) UpdateStatus(status int) error {
	c.Status = status
	if err := Set(c); err != nil {
		return fmt.Errorf("更新客户端证书状态失败: %v", err)
	}
	return nil
}

// 检查并更新证书状态为过期
func (c *ClientCertData) CheckAndUpdateStatus() error {
	if c.Status != CertStatusExpired && time.Now().After(c.NotAfter) {
		if err := c.UpdateStatus(CertStatusExpired); err != nil {
			return fmt.Errorf("更新证书状态为过期失败: %v", err)
		}
		base.Info("检测到证书过期，已更新状态:", c.Username)
	}
	return nil
}

// 获取客户端证书列表
func GetClientCertList(pageSize int, pageIndex int) ([]ClientCertData, int64, error) {
	var certs []ClientCertData
	session := GetXdb().NewSession()
	defer session.Close()
	total, err := FindAndCount(session, &certs, pageSize, pageIndex)
	if err != nil {
		return nil, 0, fmt.Errorf("获取客户端证书列表失败: %v", err)
	}
	return certs, total, nil
}

// 获取客户端证书
func GetClientCert(username string) (*ClientCertData, error) {
	clientCert := &ClientCertData{
		Username: username,
	}
	err := One("Username", username, clientCert)
	return clientCert, err
}

// 生成客户端 CA 证书
func GenerateClientCA() error {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	template := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "AnyLink Client CA"},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour * 24 * 365 * 10), // 10年有效期
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return err
	}

	// 写入 CA 证书文件
	certOut, err := os.OpenFile(base.Cfg.ClientCertCAFile, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer certOut.Close()
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})

	// 写入 CA 私钥文件
	keyOut, err := os.OpenFile(base.Cfg.ClientCertCAKeyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer keyOut.Close()
	return pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
}

// 生成客户端证书并保存到数据库
func GenerateClientCert(username, groupname string) (*ClientCertData, error) {
	// 检查是否已存在证书记录
	_, err := GetClientCert(username)
	if err != nil {
		if !errors.Is(err, ErrNotFound) {
			return nil, fmt.Errorf("获取用户证书失败: %v", err)
		}
	} else {
		// 用户已有证书记录，不允许重复生成
		return nil, fmt.Errorf("用户 %s 已存在证书，请先删除现有证书", username)
	}

	// 确保客户端 CA 已加载
	if err := LoadClientCA(); err != nil {
		return nil, fmt.Errorf("无法加载客户端 CA: %v", err)
	}

	// 生成客户端私钥
	clientKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	// 创建客户端证书模板
	template := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject: pkix.Name{
			CommonName:         username,
			OrganizationalUnit: []string{groupname},
			Organization:       []string{"AnyLink VPN"},
			Country:            []string{"CN"},
			Province:           []string{"Beijing"},
			Locality:           []string{"Beijing"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour * 24 * 365), // 1年有效期
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		DNSNames:              []string{username},
		BasicConstraintsValid: true,
		IsCA:                  false,
	}

	// 签发客户端证书
	certDER, err := x509.CreateCertificate(rand.Reader, &template, clientCACert, &clientKey.PublicKey, clientCAKey)
	if err != nil {
		return nil, err
	}

	// 编码证书和私钥为 PEM 格式
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(clientKey)})

	// 保存到数据库
	clientCertData := &ClientCertData{
		Username:     username,
		GroupName:    groupname,
		Certificate:  string(certPEM),
		PrivateKey:   string(keyPEM),
		SerialNumber: template.SerialNumber.String(),
		NotAfter:     template.NotAfter,
		CreatedAt:    time.Now(),
		Status:       CertStatusActive, // 初始状态为有效
	}

	if err := clientCertData.Save(); err != nil {
		return nil, fmt.Errorf("保存客户端证书失败: %v", err)
	}

	return clientCertData, nil
}

// 生成 PKCS#12 格式证书文件
func GenerateClientP12FromDB(username string, password string) ([]byte, error) {
	// 从数据库获取证书
	clientCert, err := GetClientCert(username)
	if err != nil {
		return nil, err
	}
	// 检查并更新证书状态
	if err := clientCert.CheckAndUpdateStatus(); err != nil {
		base.Error("检查并更新证书状态失败:", err)
	}
	// 检查证书状态
	if clientCert.GetStatus() != CertStatusActive {
		return nil, fmt.Errorf("用户 %s 的证书状态为：%s", username, clientCert.GetStatusText())
	}

	// 确保客户端 CA 已加载
	if err := LoadClientCA(); err != nil {
		return nil, fmt.Errorf("无法加载客户端 CA: %v", err)
	}

	// 解析证书和私钥
	certBlock, _ := pem.Decode([]byte(clientCert.Certificate))
	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, err
	}

	keyBlock, _ := pem.Decode([]byte(clientCert.PrivateKey))
	key, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil {
		return nil, err
	}

	// 打包为 .p12 格式
	p12Data, err := pkcs12.Modern.Encode(key, cert, []*x509.Certificate{clientCACert}, password)
	if err != nil {
		return nil, err
	}

	return p12Data, nil
}

// 验证客户端证书
func ValidateClientCert(cert *x509.Certificate, userAgent string) bool {
	// 获取用户和证书信息
	user := &User{
		Username: cert.Subject.CommonName,
	}
	err := One("Username", user.Username, user)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			base.Error("证书验证失败：用户不存在", cert.Subject.CommonName)
		} else {
			base.Error("证书验证失败：查询用户失败:", err)
		}
		return false
	}

	// 检查用户状态是否启用
	if user.Status != 1 {
		base.Error("证书验证失败：用户已禁用:", user.Username)
		return false
	}

	// 获取客户端证书记录
	clientCertData, err := GetClientCert(user.Username)
	if err != nil {
		base.Error("证书验证失败：获取客户端证书失败:", err)
		return false
	}

	if clientCertData.GroupName != cert.Subject.OrganizationalUnit[0] {
		base.Error("证书验证失败：证书组名与用户组名不匹配")
		return false
	}

	// 检查证书状态
	if clientCertData.GetStatus() != CertStatusActive {
		base.Error("证书验证失败：证书状态为", clientCertData.GetStatusText())
		return false
	}

	// 检查证书是否过期
	if time.Now().After(cert.NotAfter) {
		base.Error("证书验证失败：证书已过期:", cert.NotAfter)
		return false
	}

	// 验证证书指纹
	storedCertBlock, _ := pem.Decode([]byte(clientCertData.Certificate))
	storedCert, err := x509.ParseCertificate(storedCertBlock.Bytes)
	if err != nil {
		base.Error("证书验证失败：解析存储证书失败:", err)
		return false
	}

	// 比较证书的完整内容
	if !bytes.Equal(cert.Raw, storedCert.Raw) {
		base.Error("证书验证失败：证书内容不匹配")
		return false
	}

	// 验证证书链
	verifyOptions := x509.VerifyOptions{
		Roots:     LoadClientCAPool(),
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	if _, err := cert.Verify(verifyOptions); err != nil {
		base.Error("证书验证失败：证书链验证失败:", err)
		return false
	}
	// 检查扩展密钥用途
	hasClientAuth := slices.Contains(cert.ExtKeyUsage, x509.ExtKeyUsageClientAuth)
	if !hasClientAuth {
		base.Error("证书验证失败：证书缺少客户端认证扩展")
		return false
	}

	return true
}

// 加载客户端 CA 证书池
func LoadClientCAPool() *x509.CertPool {
	if err := LoadClientCA(); err != nil {
		return nil
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AddCert(clientCACert)
	return caCertPool
}

// 加载客户端 CA 证书和私钥
func LoadClientCA() error {
	caMutex.Lock()
	defer caMutex.Unlock()

	// 如果证书已经加载到内存中，则直接返回
	if clientCACert != nil && clientCAKey != nil {
		return nil
	}

	caCertPEM, readErr := os.ReadFile(base.Cfg.ClientCertCAFile)
	if readErr != nil {
		base.Warn("无法读取客户端 CA 证书,请初始化CA:", readErr)
		return fmt.Errorf("无法读取客户端 CA 证书,请初始化CA: %w", readErr)
	}
	caKeyPEM, readErr := os.ReadFile(base.Cfg.ClientCertCAKeyFile)
	if readErr != nil {
		return fmt.Errorf("无法读取客户端 CA 私钥: %w", readErr)
	}

	caCertBlock, _ := pem.Decode(caCertPEM)
	if caCertBlock == nil {
		return errors.New("无法解析客户端 CA 证书 PEM 块")
	}

	var parseErr error
	clientCACert, parseErr = x509.ParseCertificate(caCertBlock.Bytes)
	if parseErr != nil {
		return fmt.Errorf("无法解析客户端 CA 证书: %w", parseErr)
	}

	caKeyBlock, _ := pem.Decode(caKeyPEM)
	if caKeyBlock == nil {
		return errors.New("无法解析客户端 CA 私钥 PEM 块")
	}

	var parseKeyErr error
	clientCAKey, parseKeyErr = x509.ParsePKCS1PrivateKey(caKeyBlock.Bytes)
	if parseKeyErr != nil {
		// 解析为PKCS8
		pkcs8Key, pkcs8Err := x509.ParsePKCS8PrivateKey(caKeyBlock.Bytes)
		if pkcs8Err != nil {
			return fmt.Errorf("无法解析客户端 CA 私钥 (PKCS1 or PKCS8): %w", parseKeyErr)
		}
		var ok bool
		clientCAKey, ok = pkcs8Key.(*rsa.PrivateKey)
		if !ok {
			return errors.New("解析私钥成功，但不是 RSA 类型")
		}
	}

	return nil
}
