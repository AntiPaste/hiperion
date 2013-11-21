package main

import (
	//"fmt"
	"github.com/fzzy/radix/redis"
	"github.com/gorilla/sessions"
	"github.com/gorilla/mux"
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
	session, _ = store.Get(req, "hiperion")
	if req.FormValue("name") != "" {
		session.Values["name"] = req.FormValue("name")
		session.Save(req, w)
	}
	homeTemplate := template.Must(template.ParseFiles("home.html"))
	homeTemplate.Execute(w, struct {
		Name string
	}{
		Name: session.Values["name"]
	})
}

func staticHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		http.Error(w, "Invalid method", 405)
	}

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

	if err := http.ListenAndServe(":1234", router); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
