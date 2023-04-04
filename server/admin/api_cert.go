package admin

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/bjdgyc/anylink/base"
	"github.com/bjdgyc/anylink/dbdata"
	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
)

type LeGoClient struct {
	mutex  sync.Mutex
	Client *lego.Client
	dbdata.LegoUserData
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
	if tlscert, _, err := ParseCert(); err != nil {
		return
	} else {
		dbdata.TLSCert = tlscert
	}
	RespSucess(w, "上传成功")
}
func GetCertSetting(w http.ResponseWriter, r *http.Request) {
	data := &dbdata.SettingLetsEncrypt{}
	if err := dbdata.SettingGet(data); err != nil {
		RespError(w, RespInternalErr, err)
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
	err = json.Unmarshal(body, config)
	if err != nil {
		RespError(w, RespInternalErr, err)
		return
	}
	if err := dbdata.SettingSet(config); err != nil {
		RespError(w, RespInternalErr, err)
		return
	}
	client := LeGoClient{}
	if err := client.NewClient(config); err != nil {
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
	_, certtime, err := ParseCert()
	if err != nil {
		base.Error(err)
		return
	}
	if certtime.AddDate(0, 0, -7).Before(time.Now()) {
		config := &dbdata.SettingLetsEncrypt{}
		if err := dbdata.SettingGet(config); err != nil {
			base.Error(err)
			return
		}
		if config.Domain == "" {
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
	}
	base.Info(fmt.Sprintf("证书过期时间：%s", certtime.Local().Format("2006-1-2 15:04:05")))
}

func (c *LeGoClient) NewClient(l *dbdata.SettingLetsEncrypt) error {
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
	Provider, err := dbdata.GetDNSProvider(l)
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
	if tlscert, _, err := ParseCert(); err != nil {
		return err
	} else {
		dbdata.TLSCert = tlscert
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

func ParseCert() (*tls.Certificate, *time.Time, error) {
	cert, err := tls.LoadX509KeyPair(base.Cfg.CertFile, base.Cfg.CertKey)
	if err != nil {
		return nil, nil, err
	}
	parseCert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return nil, nil, err
	}
	certtime := parseCert.NotAfter
	return &cert, &certtime, nil
}

// func Scrypt(passwd string) string {
// 	salt := []byte{0xc8, 0x28, 0xf2, 0x58, 0xa7, 0x6a, 0xad, 0x7b}
// 	hashPasswd, err := scrypt.Key([]byte(passwd), salt, 1<<15, 8, 1, 32)
// 	if err != nil {
// 		return err.Error()
// 	}
// 	return base64.StdEncoding.EncodeToString(hashPasswd)
// }
