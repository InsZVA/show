package p2p

import (
	"fmt"
	"sync"
	"testing"
)

func TestManager(t *testing.T) {
	go Manager()
	var p [20]*Point
	wp := sync.WaitGroup{}
	for i := 0; i < 20; i++ {
		p[i] = NewPoint(nil, func(self, pair *Point) {
			fmt.Println(self, pair)
			wp.Done()
		})
		p[i].RandomPair()
		wp.Add(1)
	}
	for i := 0; i < 20; i++ {

	}
	wp.Wait()

}
