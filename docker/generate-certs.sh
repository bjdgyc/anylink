#!/bin/sh

mkdir -p /ssl

OUTPUT_FILENAME="vpn.xx.com"

printf "[req]
prompt                  = no
default_bits            = 4096
default_md              = sha256
encrypt_key             = no
string_mask             = utf8only

distinguished_name      = cert_distinguished_name
req_extensions          = req_x509v3_extensions
x509_extensions         = req_x509v3_extensions

[ cert_distinguished_name ]
C  = CN
ST = BJ
L  = BJ
O  = xx.com
OU = xx.com
CN = xx.com

[req_x509v3_extensions]
basicConstraints        = critical,CA:true
subjectKeyIdentifier    = hash
keyUsage                = critical,digitalSignature,keyCertSign,cRLSign #,keyEncipherment
extendedKeyUsage        = critical,serverAuth #, clientAuth
subjectAltName          = @alt_names

[alt_names]
DNS.1 = xx.com
DNS.2 = *.xx.com

">/ssl/${OUTPUT_FILENAME}.conf

openssl req -x509 -newkey rsa:2048 -keyout /ssl/test_vpn_key.pem -out /ssl/test_vpn_cert.pem \
-days 3600 -nodes -config /ssl/${OUTPUT_FILENAME}.conf

