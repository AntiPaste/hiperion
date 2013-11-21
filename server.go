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
	SESSION_AUTHENTICATION,
	SESSION_ENCRYPTION,
)

func homeHandler(w http.ResponseWriter, req *http.Request) {
	session, err := store.Get(req, "hiperion")

	if err != nil {
		log.Fatal("session:", err)
	}

	if session.IsNew {
		session.Options.Domain = req.Host
		session.Options.Path = "/"
		session.Options.MaxAge = 86400
		session.Options.HttpOnly = false
		session.Options.Secure = true
	}

	if req.FormValue("name") != "" {
		session.Values["name"] = req.FormValue("name")
	}

	sessions.Save(req, w)

	homeTemplate := template.Must(template.ParseFiles("home.html"))
	homeTemplate.Execute(w, struct {
		Title string
		Name  string
	}{
		Title: "Home",
		Name:  session.Values["name"].(string),
	})
}

func staticHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		http.Error(w, "Invalid method", 405)
	}

	log.Println(req.URL.Path[1:])
	http.ServeFile(w, req, req.URL.Path[1:])
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/", homeHandler)
	router.HandleFunc("/static/", staticHandler)

	conn, err := redis.Dial("tcp", "127.0.0.1:6379")
	if err != nil {
		log.Fatal("redis:", err)
	}

	db = conn
	db.Cmd("SELECT", DATABASE_INDEX)

	log.Println("Server started!")

	if err := http.ListenAndServe(":80", router); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
