package main

import (
	"io/ioutil"
	"net/http"

	"github.com/inszva/show/serv"
)

func ChatHandle(w http.ResponseWriter, r *http.Request) {
	html, err := ioutil.ReadFile("./html/index.html")
	if err != nil {
		w.WriteHeader(500)
		return
	}
	w.Write(html)
}

func main() {
	http.HandleFunc("/", serv.HandleWebsocket)
	http.HandleFunc("/chat", ChatHandle)
	err := http.ListenAndServeTLS(":443", "server.pem", "server.key", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
