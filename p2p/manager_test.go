package p2p

import (
	"sync"
	"testing"
)

func TestRandomPair(t *testing.T) {
	var p [100]*Point
	for i := 0; i < 100; i++ {
		p[i] = NewPoint(nil)
	}
	wp := sync.WaitGroup{}
	f := func(pt *Point) {
		defer wp.Done()
		pt.RandomPair()
	}
	for i := 0; i < 100; i++ {
		go f(p[i])
		wp.Add(1)
	}

	wp.Wait()

}
