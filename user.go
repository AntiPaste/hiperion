package main

import (
	"fmt"
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

	user := User{
		id:       stringToInt(data["id"]),
		username: data["username"],
		hash:     data["hash"],
		salt:     data["salt"],
	}

	return user
}
