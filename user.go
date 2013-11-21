package main

import (
	"code.google.com/p/go.crypto/scrypt"
	"encoding/hex"
	"fmt"
	"log"
)

type User struct {
	ID       int
	Username string
	Salt     string
}

func hashPassword(password string, salt string) string {
	byteHash, err := scrypt.Key([]byte(password), append([]byte(PASSWORD_SALT), []byte(salt)...), 16384, 8, 1, 32)
	if err != nil {
		log.Fatal("hashPassword:", err)
	}

	hash := hex.EncodeToString(byteHash)
	return hash
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
		ID:       stringToInt(data["id"]),
		Username: data["username"],
		Salt:     data["salt"],
	}

	return user, nil
}

func getUserID(username string) (int, error) {
	log.Println(fmt.Sprintf("username:%s", username))

	reply := db.Cmd("GET", fmt.Sprintf("username:%s", username))
	if reply.Type == redis.NilReply {
		return -1, fmt.Errorf("getUserID: user not found")
	}

	data, err := reply.Str()
	if err != nil {
		log.Fatal("getUserID:", err)
	}

	userID := stringToInt(data)
	return userID, nil
}

func handleLogin(username, password string) error {
	invalid := fmt.Errorf("Invalid username or password!")

	userID, err := getUserID(username)
	if err != nil {
		return invalid
	}

	user, _ := getUser(userID)
	hash := hashPassword(password, user.Salt)

	isValid, err := db.Cmd("SISMEMBER", "hashes", hash).Bool()
	if err != nil {
		log.Fatal("handleLogin:", err)
	}

	if isValid {
		return nil
	}

	return invalid
}
