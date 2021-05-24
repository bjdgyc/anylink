module github.com/bjdgyc/anylink

go 1.16

require (
	github.com/StackExchange/wmi v0.0.0-20210224194228-fe8f1750fd46 // indirect
	github.com/asdine/storm/v3 v3.2.1
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/go-ole/go-ole v1.2.5 // indirect
	github.com/google/gopacket v1.1.19
	github.com/gorilla/mux v1.8.0
	github.com/pion/dtls/v2 v2.0.0-00010101000000-000000000000
	github.com/pion/logging v0.2.2
	github.com/shirou/gopsutil v3.21.4+incompatible
	github.com/skip2/go-qrcode v0.0.0-20200617195104-da1b6568686e
	github.com/songgao/packets v0.0.0-20160404182456-549a10cd4091
	github.com/songgao/water v0.0.0-20200317203138-2b4b6d7c09d8
	github.com/spf13/cobra v1.1.3
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.7.0
	github.com/tklauser/go-sysconf v0.3.6 // indirect
	github.com/xhit/go-simple-mail/v2 v2.9.0
	github.com/xlzd/gotp v0.0.0-20181030022105-c8557ba2c119
	go.etcd.io/bbolt v1.3.5
	golang.org/x/crypto v0.0.0-20210513164829-c07d793c2f9a
	golang.org/x/net v0.0.0-20210520170846-37e1c6afe023
	golang.org/x/time v0.0.0-20210220033141-f8bda1e9f3ba
)

replace github.com/pion/dtls/v2 => ../dtls-2.0.9
