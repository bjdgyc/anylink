package dbdata

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"testing"

	"github.com/bjdgyc/anylink/base"
	"github.com/stretchr/testify/assert"
)

func TestGenerateClientCert(t *testing.T) {
	base.Test()
	ast := assert.New(t)

	// 设置临时目录用于测试
	tempDir := t.TempDir()
	base.Cfg.ClientCertCAFile = tempDir + "/client_ca.pem"
	base.Cfg.ClientCertCAKeyFile = tempDir + "/client_ca_key.pem"

	preIpData()
	defer closeIpdata()

	// 使用 GenerateClientCA 生成 CA
	err := GenerateClientCA()
	ast.Nil(err, "生成客户端 CA 失败")

	// 创建测试组
	group := "cert-test-group"
	dns := []ValData{{Val: "8.8.8.8"}}
	g := Group{Name: group, Status: 1, ClientDns: dns}
	err = SetGroup(&g)
	ast.Nil(err)

	// 创建测试用户
	username := "cert-test-user"
	u := User{Username: username, Groups: []string{group}, Status: 1}
	err = SetUser(&u)
	ast.Nil(err)

	// 测试证书生成成功
	certData, err := GenerateClientCert(username, group)
	ast.Nil(err)
	ast.NotNil(certData)
	ast.Equal(username, certData.Username)
	ast.Equal(group, certData.GroupName)
	ast.Equal(CertStatusActive, certData.Status)
	ast.NotEmpty(certData.Certificate)
	ast.NotEmpty(certData.PrivateKey)
	ast.NotEmpty(certData.SerialNumber)

	// 测试重复生成证书失败
	_, err = GenerateClientCert(username, group)
	ast.NotNil(err)
	ast.Contains(err.Error(), "已存在证书")

	// 测试用户不属于指定组
	_, err = GenerateClientCert(username, "nonexistent-group")
	ast.NotNil(err)
	ast.Contains(err.Error(), "不属于组")

	// 测试用户不存在
	_, err = GenerateClientCert("nonexistent-user", group)
	ast.NotNil(err)
	ast.Contains(err.Error(), "用户不存在")
}
func TestCertificateAuthFlow(t *testing.T) {
	base.Test()
	ast := assert.New(t)

	preIpData()
	defer closeIpdata()

	// 设置测试环境
	group := "auth-test-group"
	username := "auth-test-user"

	// 创建组和用户
	dns := []ValData{{Val: "8.8.8.8"}}
	g := Group{Name: group, Status: 1, ClientDns: dns}
	err := SetGroup(&g)
	ast.Nil(err)

	u := User{Username: username, Groups: []string{group}, Status: 1}
	err = SetUser(&u)
	ast.Nil(err)

	// 生成证书
	certData, err := GenerateClientCert(username, group)
	ast.Nil(err)

	// 解析证书
	cert, err := parseCertFromPEM(certData.Certificate)
	ast.Nil(err)

	// 证书验证
	valid := ValidateClientCert(cert, "test-agent")
	ast.True(valid)

	// 测试证书状态变更
	certData.Status = CertStatusDisabled
	err = certData.UpdateStatus(CertStatusDisabled)
	ast.Nil(err)

	valid = ValidateClientCert(cert, "test-agent")
	ast.False(valid)
}

func TestValidateClientCert(t *testing.T) {
	base.Test()
	ast := assert.New(t)

	// 设置临时目录用于测试
	tempDir := t.TempDir()
	base.Cfg.ClientCertCAFile = tempDir + "/client_ca.pem"
	base.Cfg.ClientCertCAKeyFile = tempDir + "/client_ca_key.pem"

	preIpData()
	defer closeIpdata()

	// 初始化客户端 CA
	err := GenerateClientCA()
	ast.Nil(err, "初始化客户端 CA 失败")

	// 创建测试组
	group := "test-group"
	dns := []ValData{{Val: "8.8.8.8"}}
	g := Group{Name: group, Status: 1, ClientDns: dns}
	err = SetGroup(&g)
	ast.Nil(err)

	// 创建测试用户
	username := "test-user"
	u := User{Username: username, Groups: []string{group}, Status: 1}
	err = SetUser(&u)
	ast.Nil(err)

	// 生成客户端证书
	certData, err := GenerateClientCert(username, group)
	ast.Nil(err)
	ast.NotNil(certData)
	ast.Equal(username, certData.Username)
	ast.Equal(group, certData.GroupName)

	// 解析生成的证书
	cert, err := parseCertFromPEM(certData.Certificate)
	ast.Nil(err)
	ast.Equal(username, cert.Subject.CommonName)
	ast.Equal(group, cert.Subject.OrganizationalUnit[0])

	// 测试证书验证成功
	valid := ValidateClientCert(cert, "test-agent")
	ast.True(valid)

	// 测试用户不存在的情况
	cert.Subject.CommonName = "nonexistent-user"
	valid = ValidateClientCert(cert, "test-agent")
	ast.False(valid)

	// 测试用户被禁用的情况
	cert.Subject.CommonName = username
	u.Status = 0
	err = SetUser(&u)
	ast.Nil(err)
	valid = ValidateClientCert(cert, "test-agent")
	ast.False(valid)

	// 恢复用户状态
	u.Status = 1
	err = SetUser(&u)
	ast.Nil(err)

	// 测试证书组不匹配的情况
	cert.Subject.OrganizationalUnit[0] = "wrong-group"
	valid = ValidateClientCert(cert, "test-agent")
	ast.False(valid)

	// 测试证书状态被禁用的情况
	cert.Subject.OrganizationalUnit[0] = group
	certData.Status = CertStatusDisabled
	err = certData.Save()
	ast.Nil(err)
	valid = ValidateClientCert(cert, "test-agent")
	ast.False(valid)
}

func parseCertFromPEM(certPEM string) (*x509.Certificate, error) {
	block, _ := pem.Decode([]byte(certPEM))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}
	return x509.ParseCertificate(block.Bytes)
}
