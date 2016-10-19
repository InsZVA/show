package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/inszva/show/serv"
)

func handleDir(w http.ResponseWriter, r *http.Request) {

	url := r.RequestURI
	dir := path.Clean(url)
	parts := strings.Split(dir, ".")
	if len(parts) == 1 {
		dir = dir + "/index.html"
	}
	log.Println(dir)
	f, err := os.OpenFile("./html/"+dir, os.O_RDONLY, 0666)
	defer f.Close()
	if err != nil {
		w.WriteHeader(404)
		return
	}
	fi, err := f.Stat()
	if err != nil {
		w.WriteHeader(500)
		return
	}
	e := `"` + fi.ModTime().String() + `"`
	w.Header().Set("Etag", e)
	w.Header().Set("Cache-Control", "max-age=2592000") // 30 days
	w.Header().Set("Pragma", "Pragma")

	if match := r.Header.Get("If-None-Match"); match != "" {
		if strings.Contains(match, e) {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}
	d, err := ioutil.ReadAll(f)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	w.Write(d)
}

func main() {
	http.HandleFunc("/serv", serv.HandleWebsocket)
	http.HandleFunc("/", handleDir)
	err := http.ListenAndServeTLS(":443", "server.pem", "server.key", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
