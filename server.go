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

func homeHandler(w http.ResponseWriter, req *http.Request) {
	session, err := store.Get(req, "hiperion")

	if err != nil {
		log.Fatal("session:", err)
	}

	if session.IsNew {
		session.Options.Domain = req.Host
		session.Options.Path = "/"
		session.Options.MaxAge = 86400 * 30
		session.Options.HttpOnly = true
		session.Options.Secure = true
	}

	if req.FormValue("name") != "" {
		session.Values["name"] = req.FormValue("name")
	}

	session.Save(req, w)

	homeTemplate := template.Must(template.ParseFiles("home.html"))
	homeTemplate.Execute(w, struct {
		Title string
		Name  string
	}{
		Title: "Home",
		Name:  session.Values["name"].(string),
	})
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/", homeHandler)
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))

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
