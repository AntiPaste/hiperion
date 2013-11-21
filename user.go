package main

import (
	"code.google.com/p/go.crypto/scrypt"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/fzzy/radix/redis"
	"github.com/gorilla/sessions"
	"log"
	"strings"
)

type User struct {
	ID       int
	Username string
	Salt     string
}

func hashUsername(username string) string {
	hasher := md5.New()
	hasher.Write([]byte(username))
	hash := hex.EncodeToString(hasher.Sum(nil))

	return hash
}

func hashPassword(password string, salt string) string {
	byteSalt, err := hex.DecodeString(salt)
	if err != nil {
		log.Fatal("hashPassword:", err)
	}

	byteHash, err := scrypt.Key([]byte(password), append([]byte(PASSWORD_SALT), byteSalt...), 16384, 8, 1, 32)
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
	username = strings.ToLower(hashUsername(username))

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

func handleLogin(session *sessions.Session, username, password string) error {
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
		session.Values["userID"] = user.ID
		return nil
	}

	return invalid
}

func handleRegistration(session *sessions.Session, username, password, verifyPassword string) []error {
	var errors []error
	if username == "" {
		errors = append(errors, fmt.Errorf("Please enter a username."))
	} else {
		usernameExists, err := db.Cmd("EXISTS", fmt.Sprintf("username:%s", username)).Bool()
		if err != nil {
			log.Fatal("handleRegistration:", err)
		}

		if usernameExists {
			errors = append(errors, fmt.Errorf("That username has already been taken!"))
		}
	}

	if password == "" {
		errors = append(errors, fmt.Errorf("Please enter a password."))
	}

	if verifyPassword == "" {
		errors = append(errors, fmt.Errorf("Please enter your password again to avoid mistakes."))
	}

	if password != verifyPassword {
		errors = append(errors, fmt.Errorf("Passwords do not match!"))
	}

	if len(errors) > 0 {
		return errors
	}

	nextID, err := db.Cmd("GET", "nextid").Int()
	if err != nil {
		log.Fatal("handleRegistration:", err)
	}

	byteSalt := make([]byte, 16)
	_, err = rand.Read(byteSalt)
	if err != nil {
		log.Fatal("handleRegistration:", err)
	}

	salt := hex.EncodeToString(byteSalt)
	hash := hashPassword(password, salt)

	db.Cmd("SET", fmt.Sprintf("username:%s", strings.ToLower(hashUsername(username))), nextID)
	db.Cmd("HMSET", fmt.Sprintf("user:%d", nextID), "id", nextID, "username", username, "salt", salt)
	db.Cmd("SADD", "hashes", hash)
	db.Cmd("INCR", "nextid")

	return errors
}
