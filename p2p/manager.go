package p2p

import (
	"sync"
	"time"
)

var (
	// autoIcreament gives each point a unique id
	autoIcreament = 0
	// Points saves all points
	Points = make(map[int]*Point)
	// Mutex to protect map
	PointsMutex = sync.RWMutex{}
)

func (p *Point) RandomPair() {
	PointsMutex.RLock()
	defer PointsMutex.RUnlock()
	p.Status = P2P_POINT_PAIRING
	for p.Status == P2P_POINT_PAIRING {
		for _, pt := range Points {
			if p.Status != P2P_POINT_PAIRING {
				return
			}
			if p == pt {
				continue
			}
			if pt.Status != P2P_POINT_PAIRING {
				continue
			}
			select {
			case pt.Chan <- true:
				if pt.Status != P2P_POINT_PAIRING {
					<-pt.Chan
					continue
				}
				select {
				case p.Chan <- true:
					if p.Status != P2P_POINT_PAIRING {
						<-p.Chan
						<-pt.Chan
						return
					}
					p.Pair = pt
					pt.Pair = p
					p.Status = P2P_POINT_OFFER
					pt.Status = P2P_POINT_ANSWER
					<-p.Chan
					<-pt.Chan
				default:
					<-pt.Chan
					time.Sleep(10 * time.Millisecond)
					continue
				}
			default:
				time.Sleep(10 * time.Millisecond)
				continue
			}
		}
	}

}

/*
func (p *Point) RandomPairUsingTryLock() {
	PointsMutex.RLock()
	defer PointsMutex.RUnlock()
	p.Status = P2P_POINT_PAIRING
	for p.Status == P2P_POINT_PAIRING {
		for _, pt := range Points {
			if p == pt {
				continue
			}
			if pt.Status != P2P_POINT_PAIRING {
				continue
			}
			if pt.Mutex.TryLock() {
				if pt.Status != P2P_POINT_PAIRING {
					pt.Mutex.Unlock()
					continue
				}
				if p.Mutex.TryLock() {
					if p.Status != P2P_POINT_PAIRING {
						p.Mutex.Unlock()
						pt.Mutex.Unlock()
						return
					}
					p.Pair = pt
					pt.Pair = p
					p.Status = P2P_POINT_OFFER
					pt.Status = P2P_POINT_ANSWER
					p.Mutex.Unlock()
					pt.Mutex.Unlock()
				} else {
					pt.Mutex.Unlock()
					//time.Sleep(10 * time.Millisecond)
					continue
				}
			} else {
				//time.Sleep(10 * time.Millisecond)
				continue
			}
		}
	}

}
*/

func GC() {
	PointsMutex.RLock()
	length := len(Points)
	PointsMutex.RUnlock()
	if length%100 == 50 {
		for i := 0; i < length/2; i++ {
			for k, p := range Points {
				if p.Status == P2P_POINT_CLOSE {
					PointsMutex.Lock()
					delete(Points, k)
					PointsMutex.Unlock()
					break
				}
			}
		}

	}
}
