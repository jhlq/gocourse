package main

import (
	"fmt"
	"log"
	"net/http"
	"text/template"

	"github.com/gorilla/mux"

	"time"
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
	dbclient := getClient()
	msgs := getRecentMessages(dbclient, 30)
	data := struct {
		Name string
		Msgs []message
	}{Name: "Anon", Msgs: msgs}
	err := tm["chat"].ExecuteTemplate(w, "base", data)
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
	msg.Author = "Anonymous"
	msg.Time = time.Now().Format("2006-01-02 15:04:05")
	addChatMessage(dbclient, msg)
	http.Redirect(w, r, "/chat/", http.StatusSeeOther)
}
func main() {
	inittm()
	http.HandleFunc("/ip", serveIP)
	router := mux.NewRouter()
	router.HandleFunc("/", serveHome)
	router.HandleFunc("/greet/{name}", serveGreet)
	router.HandleFunc("/about/", serveAbout)
	router.HandleFunc("/contact/", serveContact)
	router.HandleFunc("/chat/", serveChat).Methods("GET")
	router.HandleFunc("/chat/", handleChatMessage).Methods("POST")

	router.HandleFunc("/style.css", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "style.css")
	})
	http.Handle("/", router)
	s := &http.Server{
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		Addr:         ":8080",
	}
	log.Println("Listening on port 8080")
	err := s.ListenAndServe()
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
