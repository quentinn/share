package main

import (
	"crypto/rand"
    "encoding/base64"
)




func generatePassword() string {

	b := make([]byte, 64)
	_, err := rand.Read(b)
	if err != nil {
	   panic(err)
	}

	password := base64.StdEncoding.EncodeToString(b)

	return password
 }

