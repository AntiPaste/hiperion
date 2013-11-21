package main

import (
	//"fmt"
	"github.com/fzzy/radix/redis"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"html/template"
	"log"
	"net/http"
)

var db *redis.Client
var store = sessions.NewCookieStore(
	[]byte(SESSION_AUTHENTICATION),
	[]byte(SESSION_ENCRYPTION),
)

func getSession(req *http.Request) *sessions.Session {
	session, err := store.Get(req, "session")

	if err != nil {
		log.Fatal("getSession:", err)
	}

	if session.IsNew {
		session.Options.Domain = req.Host
		session.Options.Path = "/"
		session.Options.MaxAge = 86400 * 30
		session.Options.HttpOnly = true
		session.Options.Secure = true
	}

	return session
}

func homeHandler(w http.ResponseWriter, req *http.Request) {
	homeTemplate := template.Must(template.ParseFiles("templates/base.html", "templates/index.html"))
	homeTemplate.Execute(w, struct {
		Title string
	}{
		Title: "Home",
	})
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/", homeHandler)
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))

	conn, err := redis.Dial("tcp", "127.0.0.1:6379")
	if err != nil {
		log.Fatal("connect:", err)
	}

	db = conn
	db.Cmd("SELECT", DATABASE_INDEX)

	log.Println("Server started!")

	if err := http.ListenAndServe(":80", router); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
