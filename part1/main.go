package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"time"
)

func serveIP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "RemoteAddr: "+r.RemoteAddr)
}
func serveHome(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello from home!")
}
func serveGreet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fmt.Fprintf(w, "Hello "+vars["name"])
}
func main() {
	http.HandleFunc("/ip", serveIP)
	router := mux.NewRouter()
	router.HandleFunc("/", serveHome)
	router.HandleFunc("/greet/{name}", serveGreet)
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
