package main

import (
	"log"
	"net/http"

	"github.com/inszva/show/serv"
)

func main() {
	http.HandleFunc("/serv", serv.HandleWebsocket)
	http.Handle("/", http.FileServer(http.Dir("./html/")))
	err := http.ListenAndServeTLS(":443", "server.pem", "server.key", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
