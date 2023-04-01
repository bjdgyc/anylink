package admin

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/bjdgyc/anylink/base"
	"github.com/bjdgyc/anylink/dbdata"
	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/providers/dns/alidns"
	"github.com/go-acme/lego/v4/providers/dns/cloudflare"
	"github.com/go-acme/lego/v4/providers/dns/tencentcloud"
	"github.com/go-acme/lego/v4/registration"
	"github.com/xenolf/lego/challenge"
	"golang.org/x/crypto/scrypt"
)

type LegoUser struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}
type LeGoClient struct {
	Client *lego.Client
}

func (u *LegoUser) GetEmail() string {
	return u.Email
}
func (u LegoUser) GetRegistration() *registration.Resource {
	return u.Registration
}
func (u *LegoUser) GetPrivateKey() crypto.PrivateKey {
	return u.key
}
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
	RespSucess(w, "上传成功")
}
func GetCertSetting(w http.ResponseWriter, r *http.Request) {
	data := &dbdata.SettingDnsProvider{}
	if err := dbdata.SettingGet(data); err != nil {
		RespError(w, RespInternalErr, err)
	}
	data.AliYun.APIKey = Scrypt(data.AliYun.APIKey)
	data.AliYun.SecretKey = Scrypt(data.AliYun.SecretKey)
	data.TXCloud.SecretID = Scrypt(data.TXCloud.SecretID)
	data.TXCloud.SecretKey = Scrypt(data.TXCloud.SecretKey)
	data.CfCloud.AuthKey = Scrypt(data.CfCloud.AuthKey)
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
	config := &dbdata.SettingDnsProvider{}
	err = json.Unmarshal(body, config)
	if err != nil {
		RespError(w, RespInternalErr, err)
		return
	}
	if err := dbdata.SettingSet(config); err != nil {
		RespError(w, RespInternalErr, err)
		return
	}
	client, err := NewLeGoClient(config)
	if err != nil {
		base.Error(err)
		RespError(w, RespInternalErr, fmt.Sprintf("获取证书失败:%v", err))
		return
	}
	if err := client.GetCertificate(config.Domain); err != nil {
		base.Error(err)
		RespError(w, RespInternalErr, fmt.Sprintf("获取证书失败:%v", err))
		return
	}
	RespSucess(w, "生成证书成功")
}

func ReNewCert() {
	certtime, err := GetCerttime()
	if err != nil {
		base.Error(err)
		return
	}
	if certtime.AddDate(0, 0, -7).Before(time.Now()) {
		config := &dbdata.SettingDnsProvider{}
		if err := dbdata.SettingGet(config); err != nil {
			base.Error(err)
			return
		}
		if config.Domain == "" {
			return
		}
		if config.Renew {
			client, err := NewLeGoClient(config)
			if err != nil {
				base.Error(err)
				return
			}
			if err := client.RenewCert(base.Cfg.CertFile, base.Cfg.CertKey); err != nil {
				base.Error(err)
				return
			}
			base.Info("证书续期成功")
		}
	}
	base.Info(fmt.Sprintf("证书过期时间：%s", certtime.Local().Format("2006-1-2 15:04:05")))
}

func NewLeGoClient(d *dbdata.SettingDnsProvider) (*LeGoClient, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	legoUser := LegoUser{
		Email: d.Legomail,
		key:   privateKey,
	}
	config := lego.NewConfig(&legoUser)
	config.CADirURL = lego.LEDirectoryProduction
	config.Certificate.KeyType = certcrypto.RSA2048

	client, err := lego.NewClient(config)
	if err != nil {
		return nil, err
	}
	if _, err := client.Registration.ResolveAccountByKey(); err != nil {
		reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
		if err != nil {
			return nil, err
		}
		legoUser.Registration = reg
	}
	var Provider challenge.Provider
	switch d.Name {
	case "aliyun":
		if Provider, err = alidns.NewDNSProviderConfig(&alidns.Config{APIKey: d.AliYun.APIKey, SecretKey: d.AliYun.SecretKey, TTL: 600}); err != nil {
			return nil, err
		}
	case "txcloud":
		if Provider, err = tencentcloud.NewDNSProviderConfig(&tencentcloud.Config{SecretID: d.TXCloud.SecretID, SecretKey: d.TXCloud.SecretKey, TTL: 600}); err != nil {
			return nil, err
		}
	case "cloudflare":
		if Provider, err = cloudflare.NewDNSProviderConfig(&cloudflare.Config{AuthEmail: d.CfCloud.AuthEmail, AuthKey: d.CfCloud.AuthKey, TTL: 600}); err != nil {
			return nil, err
		}
	}
	if err := client.Challenge.SetDNS01Provider(Provider, dns01.AddRecursiveNameservers([]string{"114.114.114.114", "114.114.115.115"})); err != nil {
		return nil, err
	}
	return &LeGoClient{
		Client: client,
	}, nil
}
func (c *LeGoClient) GetCertificate(domain string) error {
	// 申请证书
	certificates, err := c.Client.Certificate.Obtain(
		certificate.ObtainRequest{
			Domains: []string{domain},
			Bundle:  true,
		})
	if err != nil {
		return err
	}
	// 保存证书
	if err := SaveCertificate(certificates); err != nil {
		return err
	}
	return nil
}

func (c *LeGoClient) RenewCert(certFile, keyFile string) error {
	cert, err := LoadCertResource(certFile, keyFile)
	if err != nil {
		return err
	}
	// 续期证书
	renewcert, err := c.Client.Certificate.Renew(certificate.Resource{
		Certificate: cert.Certificate,
		PrivateKey:  cert.PrivateKey,
	}, true, false, "")
	if err != nil {
		return err
	}
	// 保存更新证书
	if err := SaveCertificate(renewcert); err != nil {
		return err
	}
	return nil
}

func SaveCertificate(cert *certificate.Resource) error {
	err := os.WriteFile(base.Cfg.CertFile, cert.Certificate, 0600)
	if err != nil {
		return err
	}
	err = os.WriteFile(base.Cfg.CertKey, cert.PrivateKey, 0600)
	if err != nil {
		return err
	}
	return nil
}

func LoadCertResource(certFile, keyFile string) (*certificate.Resource, error) {
	cert, err := os.ReadFile(certFile)
	if err != nil {
		return nil, err
	}
	key, err := os.ReadFile(keyFile)
	if err != nil {
		return nil, err
	}
	return &certificate.Resource{
		Certificate: cert,
		PrivateKey:  key,
	}, nil
}

func GetCerttime() (*time.Time, error) {
	cert, err := tls.LoadX509KeyPair(base.Cfg.CertFile, base.Cfg.CertKey)
	if err != nil {
		return nil, err
	}
	parseCert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return nil, err
	}
	certtime := parseCert.NotAfter
	return &certtime, nil
}

func Scrypt(passwd string) string {
	salt := []byte{0xc8, 0x28, 0xf2, 0x58, 0xa7, 0x6a, 0xad, 0x7b}
	hashPasswd, err := scrypt.Key([]byte(passwd), salt, 1<<15, 8, 1, 32)
	if err != nil {
		return err.Error()
	}
	return base64.StdEncoding.EncodeToString(hashPasswd)
}
