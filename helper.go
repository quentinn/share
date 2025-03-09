package main

import (
	"crypto/rand"
    "encoding/base64"
	"strings"
)




func generatePassword() string {

	b := make([]byte, 64)
	_, err := rand.Read(b)
	if err != nil {
	   panic(err)
	}

	password := base64.StdEncoding.EncodeToString(b)
	password = strings.Replace(password, "/", "", -1)			// Remove unwanted char to avoid URL issues
	password = strings.Replace(password, "+", "", -1)			// Remove unwanted char to avoid URL issues


	return password
 }

