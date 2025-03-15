package main

import (
	"fmt"
	"os"
	"strings"
    "encoding/base64"
	"crypto/rand"
)




// Generate a secure string
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




// Delete a file or directory from filesystem
func deletePath(path string) {
	err := os.RemoveAll(path)
	if err != nil {
		fmt.Println(err)
	}
}