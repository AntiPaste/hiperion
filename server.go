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
		session.Options.Path = "/login"
		session.Options.MaxAge = 86400 * 30
		session.Options.HttpOnly = false
		session.Options.Secure = false
	}

	return session
}

func homeHandler(w http.ResponseWriter, req *http.Request) {
	session := getSession(req)
	if session.Values["userID"] == nil {
		http.Redirect(w, req, "/login", http.StatusSeeOther)
		return
	}

	homeTemplate := template.Must(template.ParseFiles("templates/base.html", "templates/index.html"))
	homeTemplate.Execute(w, struct {
		CSS []string
		JS  []string
	}{
		CSS: []string{},
		JS:  []string{},
	})
}

func loginHandler(w http.ResponseWriter, req *http.Request) {
	session := getSession(req)
	if session.Values["userID"] != nil {
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}

	if req.Method == "POST" {
		username := req.PostFormValue("username")
		password := req.PostFormValue("password")
		err := handleLogin(username, password)

		if err != nil {
			session.AddFlash(err.Error(), "errors")
			session.AddFlash("test", "errors")
			x := session.Save(req, w)
			if x != nil {
				log.Println(x)
			}

			http.Redirect(w, req, "/login", http.StatusSeeOther)
			return
		}

		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}

	c := session.Save(req, w)
	if c != nil {
		log.Println(c)
	}

	errors := session.Flashes("errors")
	log.Println(errors)

	loginTemplate := template.Must(template.ParseFiles("templates/base.html", "templates/login.html"))
	loginTemplate.Execute(w, struct {
		CSS    []string
		JS     []string
		Errors []interface{}
	}{
		CSS:    []string{},
		JS:     []string{},
		Errors: errors,
	})
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/", homeHandler)
	router.HandleFunc("/login", loginHandler)
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
