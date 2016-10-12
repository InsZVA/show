package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type WS struct {
	Conn *websocket.Conn
	Type int
}

var (
	upgrader   = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	waitQueue  = make([]string, 0)
	waitSocket = make(map[string]WS)
	pairSocket = make(map[string]string)
)

func HelloServer(w http.ResponseWriter, req *http.Request) {
	log.Println(req.RemoteAddr)
	data, err := ioutil.ReadFile("./wait.html")
	if err != nil {
		w.WriteHeader(404)
		return
	}
	w.Write(data)
}

func ChatHandle(w http.ResponseWriter, req *http.Request) {
	log.Println(req.RemoteAddr)
	data, err := ioutil.ReadFile("./1.html")
	if err != nil {
		w.WriteHeader(404)
		return
	}
	w.Write(data)
}

func BroadCast() {
	resp := make(map[string]interface{})
	for _, m := range waitSocket {
		resp["type"] = "update"
		resp["num"] = len(waitQueue)
		respMsg, _ := json.Marshal(resp)
		err := m.Conn.WriteMessage(m.Type, respMsg)
		if err != nil {
			log.Println("write:", err)
		}
	}
}

func WaitQueueHandle(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	log.Println(r.RemoteAddr, "Connet to server")
	defer func() {
		c.Close()
		delete(waitSocket, r.RemoteAddr)
	}()
	for {
		var resp = make(map[string]interface{})
		respMsg := []byte("{}")
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		var jsonData map[string]interface{}
		err = json.Unmarshal(message, &jsonData)
		if err != nil {
			log.Println(err)
			continue
		}
		log.Printf("recv: %s", jsonData)
		typeMsg, ok := jsonData["type"].(string)
		if !ok {
			log.Println("type missing")
			continue
		}
		if typeMsg == "ready" {
			waitSocket[r.RemoteAddr] = WS{
				Conn: c,
				Type: mt,
			}
			resp["type"] = "update"
			resp["num"] = len(waitQueue)
			respMsg, _ = json.Marshal(resp)
		} else if typeMsg == "join" {
			if len(waitQueue) == 0 {
				waitQueue = append(waitQueue, r.RemoteAddr)
				BroadCast()
				continue
			} else {
				pair := waitQueue[0]
				waitQueue = append([]string{}, waitQueue[1:]...)
				BroadCast()
				resp["type"] = "start"
				resp["remoteIp"] = pair
				pairSocket[pair] = r.RemoteAddr
				pairSocket[r.RemoteAddr] = pair
				respMsg, _ = json.Marshal(resp)
				c.WriteMessage(mt, respMsg)

				resp_t := make(map[string]interface{})
				resp_t["type"] = "start2"
				resp_t["remoteIp"] = pair
				respMsg_t, _ := json.Marshal(resp_t)
				waitSocket[pair].Conn.WriteMessage(mt, respMsg_t)

				continue
			}
		} else if typeMsg == "__ice_candidate" || typeMsg == "__offer" || typeMsg == "__answer" {
			waitSocket[pairSocket[r.RemoteAddr]].Conn.WriteMessage(mt, message)
			continue
		}
		err = c.WriteMessage(mt, respMsg)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func main() {
	http.HandleFunc("/", HelloServer)
	http.HandleFunc("/queue", WaitQueueHandle)
	http.HandleFunc("/chat", ChatHandle)
	err := http.ListenAndServeTLS(":443", "server.pem", "server.key", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
