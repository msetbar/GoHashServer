package main

import (
	"crypto/sha512"
	"encoding/base64"
)

func GetHashedPassword(password string) string {

	sha_512 := sha512.New()
	sha_512.Write([]byte(password))

	return base64.StdEncoding.EncodeToString(sha_512.Sum(nil))
}
