package dbdata

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"

	"github.com/go-acme/lego/v4/providers/dns/alidns"
	"github.com/go-acme/lego/v4/providers/dns/cloudflare"
	"github.com/go-acme/lego/v4/providers/dns/tencentcloud"
	"github.com/go-acme/lego/v4/registration"
	"github.com/xenolf/lego/challenge"
)

var TLSCert *tls.Certificate

type SettingLetsEncrypt struct {
	// LegoUser LegoUser
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
