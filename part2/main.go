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

func inittm() {
	tm = make(map[string]*template.Template)
	tm["home"] = template.Must(template.ParseFiles("templates/home.gohtml", "templates/base.gohtml"))
	tm["about"] = template.Must(template.ParseFiles("templates/about.gohtml", "templates/base.gohtml"))
	tm["contact"] = template.Must(template.ParseFiles("templates/contact.gohtml", "templates/base.gohtml"))
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
func main() {
	inittm()
	http.HandleFunc("/ip", serveIP)
	router := mux.NewRouter()
	router.HandleFunc("/", serveHome)
	router.HandleFunc("/greet/{name}", serveGreet)
	router.HandleFunc("/about/", serveAbout)
	router.HandleFunc("/contact/", serveContact)
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
