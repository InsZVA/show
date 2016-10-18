package serv

import (
	"log"
	"net/http"
	"testing"
)

func TestHandleWebsocket(t *testing.T) {
	http.HandleFunc("/", HandleWebsocket)
	err := http.ListenAndServeTLS(":443", "../server.pem", "../server.key", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
