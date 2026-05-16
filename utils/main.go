package main

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	password := "pass1234"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	fmt.Printf(
		"INSERT INTO admin_credentials (login, password_hash) VALUES ('admin', '%s');\n",
		hash,
	)
}
