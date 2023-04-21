package dbdata

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/pion/dtls/v2/pkg/crypto/selfsign"

	"github.com/bjdgyc/anylink/base"
	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/providers/dns/alidns"
	"github.com/go-acme/lego/v4/providers/dns/cloudflare"
	"github.com/go-acme/lego/v4/providers/dns/tencentcloud"
	"github.com/go-acme/lego/v4/registration"
)

var nameToCertificate = make(map[string]*tls.Certificate)

var tempCert *tls.Certificate

func init() {
	c, _ := selfsign.GenerateSelfSignedWithDNS("localhost")
	tempCert = &c
}

type SettingLetsEncrypt struct {
	Domain   string `json:"domain"`
	Legomail string `json:"legomail"`
	Name     string `json:"name"`
	Renew    bool   `json:"renew"`
	DNSProvider
}

type DNSProvider struct {
	AliYun struct {
		APIKey    string `json:"apiKey"`
		SecretKey string `json:"secretKey"`
	} `json:"aliyun"`

	TXCloud struct {
		SecretID  string `json:"secretId"`
		SecretKey string `json:"secretKey"`
	} `json:"txcloud"`
	CfCloud struct {
		AuthEmail string `json:"authEmail"`
		AuthKey   string `json:"authKey"`
	} `json:"cfcloud"`
}
type LegoUserData struct {
	Email        string                 `json:"email"`
	Registration *registration.Resource `json:"registration"`
	Key          []byte                 `json:"key"`
}
type LegoUser struct {
	Email        string
	Registration *registration.Resource
	Key          *ecdsa.PrivateKey
}

type LeGoClient struct {
	mutex  sync.Mutex
	Client *lego.Client
	Cert   *certificate.Resource
	LegoUserData
}

func GetDNSProvider(l *SettingLetsEncrypt) (Provider challenge.Provider, err error) {
	switch l.Name {
	case "aliyun":
		if Provider, err = alidns.NewDNSProviderConfig(&alidns.Config{APIKey: l.DNSProvider.AliYun.APIKey, SecretKey: l.DNSProvider.AliYun.SecretKey, TTL: 600}); err != nil {
			return
		}
	case "txcloud":
		if Provider, err = tencentcloud.NewDNSProviderConfig(&tencentcloud.Config{SecretID: l.DNSProvider.TXCloud.SecretID, SecretKey: l.DNSProvider.TXCloud.SecretKey, TTL: 600}); err != nil {
			return
		}
	case "cloudflare":
		if Provider, err = cloudflare.NewDNSProviderConfig(&cloudflare.Config{AuthEmail: l.DNSProvider.CfCloud.AuthEmail, AuthKey: l.DNSProvider.CfCloud.AuthKey, TTL: 600}); err != nil {
			return
		}
	}
	return
}
func (u *LegoUser) GetEmail() string {
	return u.Email
}
func (u LegoUser) GetRegistration() *registration.Resource {
	return u.Registration
}
func (u *LegoUser) GetPrivateKey() crypto.PrivateKey {
	return u.Key
}

func (l *LegoUserData) SaveUserData(u *LegoUser) error {
	key, err := x509.MarshalECPrivateKey(u.Key)
	if err != nil {
		return err
	}
	l.Email = u.Email
	l.Registration = u.Registration
	l.Key = key
	if err := SettingSet(l); err != nil {
		return err
	}
	return nil
}

func (l *LegoUserData) GetUserData(d *SettingLetsEncrypt) (*LegoUser, error) {
	if err := SettingGet(l); err != nil {
		return nil, err
	}
	if l.Email != "" {
		key, err := x509.ParseECPrivateKey(l.Key)
		if err != nil {
			return nil, err
		}
		return &LegoUser{
			Email:        l.Email,
			Registration: l.Registration,
			Key:          key,
		}, nil
	}
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	return &LegoUser{
		Email: d.Legomail,
		Key:   privateKey,
	}, nil
}
func ReNewCert() {
	_, certtime, err := ParseCert()
	if err != nil {
		base.Error(err)
		return
	}
	if certtime.AddDate(0, 0, -7).Before(time.Now()) {
		config := &SettingLetsEncrypt{}
		if err := SettingGet(config); err != nil {
			base.Error(err)
			return
		}
		if config.Renew {
			client := &LeGoClient{}
			if err := client.NewClient(config); err != nil {
				base.Error(err)
				return
			}
			if err := client.RenewCert(base.Cfg.CertFile, base.Cfg.CertKey); err != nil {
				base.Error(err)
				return
			}
			base.Info("证书续期成功")
		}
	} else {
		base.Info(fmt.Sprintf("证书过期时间：%s", certtime.Local().Format("2006-1-2 15:04:05")))
	}
}

func (c *LeGoClient) NewClient(l *SettingLetsEncrypt) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	legouser, err := c.GetUserData(l)
	if err != nil {
		return err
	}
	config := lego.NewConfig(legouser)
	config.CADirURL = lego.LEDirectoryProduction
	config.Certificate.KeyType = certcrypto.RSA2048

	client, err := lego.NewClient(config)
	if err != nil {
		return err
	}
	Provider, err := GetDNSProvider(l)
	if err != nil {
		return err
	}
	if err := client.Challenge.SetDNS01Provider(Provider, dns01.AddRecursiveNameservers([]string{"114.114.114.114", "114.114.115.115"})); err != nil {
		return err
	}
	if legouser.Registration == nil {
		reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
		if err != nil {
			return err
		}
		legouser.Registration = reg
		c.SaveUserData(legouser)
	}
	c.Client = client
	return nil
}

func (c *LeGoClient) GetCert(domain string) error {
	// 申请证书
	certificates, err := c.Client.Certificate.Obtain(
		certificate.ObtainRequest{
			Domains: []string{domain},
			Bundle:  true,
		})
	if err != nil {
		return err
	}
	c.Cert = certificates
	// 保存证书
	if err := c.SaveCert(); err != nil {
		return err
	}
	return nil
}

func (c *LeGoClient) RenewCert(certFile, keyFile string) error {
	cert, err := os.ReadFile(certFile)
	if err != nil {
		return err
	}
	key, err := os.ReadFile(keyFile)
	if err != nil {
		return err
	}
	// 续期证书
	renewcert, err := c.Client.Certificate.Renew(certificate.Resource{
		Certificate: cert,
		PrivateKey:  key,
	}, true, false, "")
	if err != nil {
		return err
	}
	c.Cert = renewcert
	// 保存更新证书
	if err := c.SaveCert(); err != nil {
		return err
	}
	return nil
}

func (c *LeGoClient) SaveCert() error {
	err := os.WriteFile(base.Cfg.CertFile, c.Cert.Certificate, 0600)
	if err != nil {
		return err
	}
	err = os.WriteFile(base.Cfg.CertKey, c.Cert.PrivateKey, 0600)
	if err != nil {
		return err
	}
	if tlscert, _, err := ParseCert(); err != nil {
		return err
	} else {
		LoadCertificate(tlscert)
	}
	return nil
}

func ParseCert() (*tls.Certificate, *time.Time, error) {
	_, errCert := os.Stat(base.Cfg.CertFile)
	_, errKey := os.Stat(base.Cfg.CertKey)
	if os.IsNotExist(errCert) || os.IsNotExist(errKey) {
		err := PrivateCert()
		if err != nil {
			return nil, nil, err
		}
	}
	cert, err := tls.LoadX509KeyPair(base.Cfg.CertFile, base.Cfg.CertKey)
	if err != nil || errors.Is(err, os.ErrNotExist) {
		PrivateCert()
		return nil, nil, err
	}
	parseCert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return nil, nil, err
	}
	return &cert, &parseCert.NotAfter, nil
}

func PrivateCert() error {
	// 创建一个RSA密钥对
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	pub := &priv.PublicKey

	// 生成一个自签名证书
	template := x509.Certificate{
		SerialNumber:          big.NewInt(1658),
		Subject:               pkix.Name{CommonName: "localhost"},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour * 24 * 365),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           []net.IP{},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, pub, priv)
	if err != nil {
		return err
	}

	// 将证书编码为PEM格式并将其写入文件
	certOut, _ := os.OpenFile(base.Cfg.CertFile, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0600)
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certOut.Close()

	// 将私钥编码为PEM格式并将其写入文件
	keyOut, _ := os.OpenFile(base.Cfg.CertKey, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	keyOut.Close()
	cert, err := tls.LoadX509KeyPair(base.Cfg.CertFile, base.Cfg.CertKey)
	if err != nil {
		return err
	}
	LoadCertificate(&cert)
	return nil
}

func getTempCertificate() (*tls.Certificate, error) {
	var err error
	var cert tls.Certificate
	if tempCert == nil {
		cert, err = selfsign.GenerateSelfSignedWithDNS("localhost")
		tempCert = &cert
	}
	return tempCert, err
}

func GetCertificateBySNI(commonName string) (*tls.Certificate, error) {
	// Copy from tls.Config getCertificate()
	name := strings.ToLower(commonName)
	if cert, ok := nameToCertificate[name]; ok {
		return cert, nil
	}
	if len(name) > 0 {
		labels := strings.Split(name, ".")
		labels[0] = "*"
		wildcardName := strings.Join(labels, ".")
		if cert, ok := nameToCertificate[wildcardName]; ok {
			return cert, nil
		}
	}
	// 默认证书 兼容不支持 SNI 的客户端
	if cert, ok := nameToCertificate["default"]; ok {
		return cert, nil
	}

	return getTempCertificate()
}

func LoadCertificate(cert *tls.Certificate) {
	buildNameToCertificate(cert)
}

// Copy from tls.Config BuildNameToCertificate()
func buildNameToCertificate(cert *tls.Certificate) {
	// 设置默认证书
	nameToCertificate["default"] = cert

	x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return
	}
	startTime := x509Cert.NotBefore.String()
	expiredTime := x509Cert.NotAfter.String()
	if x509Cert.Subject.CommonName != "" && len(x509Cert.DNSNames) == 0 {
		commonName := x509Cert.Subject.CommonName
		fmt.Printf("┏ Load Certificate: %s\n", commonName)
		fmt.Printf("┠╌╌ Start Time:     %s\n", startTime)
		fmt.Printf("┖╌╌ Expired Time:   %s\n", expiredTime)
		nameToCertificate[commonName] = cert
	}
	for _, san := range x509Cert.DNSNames {
		fmt.Printf("┏ Load Certificate: %s\n", san)
		fmt.Printf("┠╌╌ Start Time:     %s\n", startTime)
		fmt.Printf("┖╌╌ Expired Time:   %s\n", expiredTime)
		nameToCertificate[san] = cert
	}
}

// func Scrypt(passwd string) string {
// 	salt := []byte{0xc8, 0x28, 0xf2, 0x58, 0xa7, 0x6a, 0xad, 0x7b}
// 	hashPasswd, err := scrypt.Key([]byte(passwd), salt, 1<<15, 8, 1, 32)
// 	if err != nil {
// 		return err.Error()
// 	}
// 	return base64.StdEncoding.EncodeToString(hashPasswd)
// }
