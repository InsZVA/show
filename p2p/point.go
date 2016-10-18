package p2p

import (
	"errors"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	WS_CLOSE_ERROR_CODES = []int{000, 1001, 1002, 1003, 1004, 1005, 1006,
		1007, 1008, 1009, 1010, 1011, 1012, 1013, 1015}
)

const (
	P2P_POINT_READY = iota
	P2P_POINT_CLOSE
	P2P_POINT_PAIRING
	P2P_POINT_OFFER
	P2P_POINT_ANSWER
	P2P_POINT_CHATING
)

var (
	P2P_POINT_CLOSED_ERROR = errors.New("The point has been closed.")
)

// Point struct stands for a specified client
type Point struct {
	Conn     *websocket.Conn
	Mutex    sync.Mutex // Protect websocket conn
	Status   int
	Pair     *Point
	Chan     chan bool               // When pairing, Chan is needed to avoid conflicts
	ListChan chan bool               // Protect the list when mapping When a chain was locked, its value cannot be read
	OnPair   func(self, pair *Point) //Event handler when pair is ok
}

// Push data to client, if client is going away, the status will be P2P_POINT_CLOSE.
func (p *Point) Push(data map[string]interface{}) error {
	if p.Status == P2P_POINT_CLOSE {
		return P2P_POINT_CLOSED_ERROR
	}
	p.Mutex.Lock()
	err := p.Conn.WriteJSON(data)
	p.Mutex.Unlock()
	if websocket.IsCloseError(err, WS_CLOSE_ERROR_CODES...) {
		p.Status = P2P_POINT_CLOSE
	}
	return err
}

// Pull data from client, if client is going away, the status will be P2P_POINT_CLOSE.
func (p *Point) Pull() (map[string]interface{}, error) {
	data := make(map[string]interface{})
	if p.Status == P2P_POINT_CLOSE {
		return nil, P2P_POINT_CLOSED_ERROR
	}
	err := p.Conn.ReadJSON(&data)
	if websocket.IsCloseError(err, WS_CLOSE_ERROR_CODES...) {
		p.Status = P2P_POINT_CLOSE
	}
	return data, err
}

// Create a new point with a special websocket connection.
func NewPoint(conn *websocket.Conn, onpair func(self, pair *Point)) (p *Point) {
	p = &Point{
		Conn:     conn,
		Status:   P2P_POINT_READY,
		Chan:     make(chan bool, 1),
		ListChan: make(chan bool, 1),
		OnPair:   onpair,
	}
	b := Points.Back()
	if b != nil {
		b.Value.(*Point).ListChan <- true
		Points.PushBack(p)
		<-b.Value.(*Point).ListChan
	} else {
		Points.PushBack(p)
	}

	if conn == nil {
		return
	}
	log.Println(conn.RemoteAddr().String(), "connected to server")
	return
}
