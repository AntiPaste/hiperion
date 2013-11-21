package main

import (
	"fmt"
	"log"
)

type User struct {
	id       int
	username string
	hash     string
	salt     string
}

func getUser(identifier int) (User, error) {
	data, err := db.Cmd("HGETALL", fmt.Sprintf("user:%d", identifier)).Hash()
	if err != nil {
		log.Fatal("getUser:", err)
	}

	if len(data) == 0 {
		return User{}, fmt.Errorf("getUser: user not found")
	}

	user := User{
		id:       stringToInt(data["id"]),
		username: data["username"],
		hash:     data["hash"],
		salt:     data["salt"],
	}

	return user, nil
}

func handleLogin(input map[string][]string) error {
	if input["username"] == "test" && input["password"] == "1234" {
		return nil
	}

	return fmt.Errorf("Invalid username or password!")
}
