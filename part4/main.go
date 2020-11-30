package main

import (
	"fmt"
	"log"
	"net/http"
	"text/template"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"time"

	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"

	uuid "github.com/satori/go.uuid"
)

var tm map[string]*template.Template
var dbclient = getClient()

func inittm() {
	tm = make(map[string]*template.Template)
	tm["home"] = template.Must(template.ParseFiles("templates/home.gohtml", "templates/base.gohtml"))
	tm["about"] = template.Must(template.ParseFiles("templates/about.gohtml", "templates/base.gohtml"))
	tm["contact"] = template.Must(template.ParseFiles("templates/contact.gohtml", "templates/base.gohtml"))
	tm["chat"] = template.Must(template.ParseFiles("templates/chat.gohtml", "templates/base.gohtml"))
}
func serveIP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "RemoteAddr: "+r.RemoteAddr)
}
func serveHome(w http.ResponseWriter, r *http.Request) {
	err := tm["home"].ExecuteTemplate(w, "base", struct{}{})
	if err != nil {
		log.Println(err)
	}
}
func serveGreet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fmt.Fprintf(w, "Hello "+vars["name"])
}
func serveAbout(w http.ResponseWriter, r *http.Request) {
	err := tm["about"].ExecuteTemplate(w, "base", struct{}{})
	if err != nil {
		log.Println(err)
	}
}
func serveContact(w http.ResponseWriter, r *http.Request) {
	err := tm["contact"].ExecuteTemplate(w, "base", struct{ ContactMethod string }{ContactMethod: "Bottled message."})
	if err != nil {
		log.Println(err)
	}
}
func serveChat(w http.ResponseWriter, r *http.Request) {
	msgs := getRecentMessages(dbclient, 30)
	data := struct {
		Name string
		Msgs []message
	}{Name: "Anon", Msgs: msgs}
	session, err := gothic.Store.Get(r, "session")
	if session.Values["Name"] != nil && session.Values["Name"] != "" {
		data.Name = session.Values["Name"].(string)
	}
	err = tm["chat"].ExecuteTemplate(w, "base", data)
	if err != nil {
		log.Println(err)
	}
}
func handleChatMessage(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Println(err)
		return
	}
	var msg message
	msg.Message = r.PostForm["message"][0]
	session, _ := gothic.Store.Get(r, "session")
	if session.Values["Name"] == nil || session.Values["Name"] == "" {
		msg.Author = "Anonymous"
	} else {
		msg.Author = session.Values["Name"].(string)
	}
	msg.Time = time.Now().Format("2006-01-02 15:04:05")
	addChatMessage(dbclient, msg)
	http.Redirect(w, r, "/chat/", http.StatusSeeOther)
}
func callback(res http.ResponseWriter, req *http.Request) {
	user, err := gothic.CompleteUserAuth(res, req)
	if err != nil {
		fmt.Println("Error in CompleteUserAuth: ", err)
		fmt.Fprintln(res, err)
		return
	}
	session, err := gothic.Store.Get(req, "session")
	if err != nil {
		fmt.Println("Error getting session: ", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	id := uuid.NewV4()
	session.Values["uuid"] = id.String()
	session.Values["Name"] = user.Name
	session.Values["Email"] = user.Email
	err = session.Save(req, res)
	if err != nil {
		fmt.Println("Error saving session: ", err)
	}
	http.Redirect(res, req, "/chat/", http.StatusSeeOther)
}
func logout(res http.ResponseWriter, req *http.Request) {
	gothic.Logout(res, req)
	session, _ := gothic.Store.Get(req, "session")
	session.Values["Email"] = ""
	session.Values["Name"] = ""
	session.Values["uuid"] = ""
	session.Save(req, res)
	res.Header().Set("Location", "/")
	res.WriteHeader(http.StatusSeeOther)
}
func authenticate(res http.ResponseWriter, req *http.Request) {
	gothic.BeginAuthHandler(res, req)
}
func main() {
	key := "abc"         // Replace with your SESSION_SECRET or similar
	maxAge := 86400 * 30 // 30 days
	isProd := false      // Set to true when serving over https

	store := sessions.NewCookieStore([]byte(key))
	store.MaxAge(maxAge)
	store.Options.Path = "/"
	store.Options.HttpOnly = true // HttpOnly should always be enabled
	store.Options.Secure = isProd

	gothic.Store = store

	goth.UseProviders(
		google.New("1032837579249-do4m81l5a362c84pr6j1d20oc7i19h7h.apps.googleusercontent.com", "yN-0gr2ta_-Q61frHWJ3uhQJ", "http://127.0.0.1:8080/auth/google/callback", "email", "profile"),
	)

	inittm()
	http.HandleFunc("/ip", serveIP)
	router := mux.NewRouter()
	router.HandleFunc("/", serveHome)
	router.HandleFunc("/greet/{name}", serveGreet)
	router.HandleFunc("/about/", serveAbout)
	router.HandleFunc("/contact/", serveContact)
	router.HandleFunc("/chat/", serveChat).Methods("GET")
	router.HandleFunc("/chat/", handleChatMessage).Methods("POST")

	router.HandleFunc("/auth/{provider}/callback", callback)
	router.HandleFunc("/logout/{provider}", logout)
	router.HandleFunc("/auth/{provider}", authenticate)

	router.HandleFunc("/style.css", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "style.css")
	})
	http.Handle("/", router)
	s := &http.Server{
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		Addr:         ":8080",
	}
	log.Println("Listening on port ", s.Addr)
	err := s.ListenAndServe()
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
