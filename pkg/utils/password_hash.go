package utils

import (
	"crypto/rand"
	"encoding/base64"
	mt "math/rand"

	"golang.org/x/crypto/bcrypt"
)

func PasswordHash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func PasswordVerify(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// $sha-256$salt-key$hash-abcd
// $sha-512$salt-key$hash-abcd
const (
	saltSize = 16
	delmiter = "$"
)

func saltSecret() (string, error) {
	rb := make([]byte, randInt(10, 100))
	_, err := rand.Read(rb)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(rb), nil
}

func randInt(min int, max int) int {
	return min + mt.Intn(max-min)
}
