package p2p

import (
	"container/list"
	"log"
	"sync/atomic"
	//"time"
)

var (
	// Points saves all points
	Points = list.New()
	// Atomic integer to control manager number
	Pairing_num int32
	Manager_num int32
	//Message to broadcast
	BroadCastMessage = make(chan map[string]interface{}, 1)
	// Chan to control manager's spinning
	Exist_pairing = make(chan bool, 1)
)

func (p *Point) RandomPair() {
	p.Status = P2P_POINT_PAIRING
	atomic.AddInt32(&Pairing_num, 1)
	select {
	case Exist_pairing <- true:
	default:
	}
}

func init() {
	go Manager()
}

// Duplicated because of starve problem
/*
func (p *Point) RandomPair() {
	panic("RandomPair() function has been duplicated")
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

}*/

func Manager() {
	atomic.AddInt32(&Manager_num, 1)
	log.Println("Manager Start")
	defer log.Println("Manager die")
	for {
		select {
		case <-Exist_pairing:
			if atomic.LoadInt32(&Pairing_num) > 1 {
				select {
				case Exist_pairing <- true:
				default:
				}
			}

			//Manager number control
			if atomic.LoadInt32(&Pairing_num) > 50 {
				go Manager()
			}
			if atomic.LoadInt32(&Pairing_num)/atomic.LoadInt32(&Manager_num) < 20 && atomic.LoadInt32(&Manager_num) > 1 {
				atomic.AddInt32(&Manager_num, -1)
				return
			}

			var p0 *Point
			l := Points.Front()
			for ; l != nil && (p0 == nil || p0.Status == P2P_POINT_PAIRING); l = l.Next() {
				p := l.Value.(*Point)
				select {
				case p.ListChan <- true:
					<-p.ListChan
				default:
					if p0 != nil {
						<-p0.Chan
					}
					goto bbreak
				}
				if p.Status == P2P_POINT_PAIRING {
					if p0 == nil {
						select {
						case p.Chan <- true:
							if p.Status != P2P_POINT_PAIRING {
								<-p.Chan
								continue
							}
							p0 = p
							continue
						default:
							continue
						}
					}
					select {
					case p.Chan <- true:
						if p0.Status != P2P_POINT_PAIRING {
							<-p0.Chan
							<-p.Chan
							p0 = nil
							goto bbreak
						}
						if p.Status != P2P_POINT_PAIRING {
							<-p.Chan
							continue
						}
						p0.Status = P2P_POINT_OFFER
						p.Status = P2P_POINT_ANSWER
						p0.Pair = p
						p.Pair = p0
						atomic.AddInt32(&Pairing_num, -2)
						<-p.Chan
						<-p0.Chan
						if atomic.LoadInt32(&Pairing_num) > 1 {
							select {
							case Exist_pairing <- true:
							default:
							}
						}
						go p0.OnPair(p0, p)
						go p.OnPair(p, p0)
						goto bbreak
					default:
						continue
					}
				}
			}
			if p0 != nil {
				<-p0.Chan
			}
		bbreak:
			p0 = nil
		case msg := <-BroadCastMessage:
			Laji := make([]*list.Element, 0)
			if msg != nil {
				for l := Points.Front(); l != nil; l = l.Next() {
					p := l.Value.(*Point)
					if p.Status != P2P_POINT_CLOSE {
						p.Push(msg)
					} else {
						Laji = append(Laji, l)
					}
				}
				for _, l := range Laji {
					var p *Point
					if l.Prev() != nil {
						p = l.Prev().Value.(*Point)
					} else {
						p = l.Value.(*Point)
					}
					select {
					case p.Chan <- true:
						log.Println("Clean Laji")
						Points.Remove(l)
						<-p.Chan
					default:
					}

				}
			}
		}
	}
	/*
		for atomic.LoadInt32(&Pairing_num) > 1 || atomic.LoadInt32((&Manager_num)) == 1 {

			//Manager number control
			if atomic.LoadInt32(&Pairing_num) > 50 {
				go Manager()
			}
			if atomic.LoadInt32(&Pairing_num)/atomic.LoadInt32(&Manager_num) < 20 && atomic.LoadInt32(&Manager_num) > 1 {
				atomic.AddInt32(&Manager_num, -1)
				return
			}

			//Pick broadcast message
			select {
			case msg := <-BroadCastMessage:
				Laji := make([]*list.Element, 0)
				if msg != nil {
					for l := Points.Front(); l != nil; l = l.Next() {
						p := l.Value.(*Point)
						if p.Status != P2P_POINT_CLOSE {
							p.Push(msg)
						} else {
							Laji = append(Laji, l)
						}
					}
					for _, l := range Laji {
						var p *Point
						if l.Prev() != nil {
							p = l.Prev().Value.(*Point)
						} else {
							p = l.Value.(*Point)
						}
						select {
						case p.Chan <- true:
							log.Println("Clean Laji")
							Points.Remove(l)
							<-p.Chan
						default:
						}

					}
				}
			default:
			}

			var p0 *Point
			l := Points.Front()
			for ; l != nil && (p0 == nil || p0.Status == P2P_POINT_PAIRING); l = l.Next() {
				p := l.Value.(*Point)
				select {
				case p.ListChan <- true:
					<-p.ListChan
				default:
					if p0 != nil {
						<-p0.Chan
					}
					goto bbreak
				}
				if p.Status == P2P_POINT_PAIRING {
					if p0 == nil {
						select {
						case p.Chan <- true:
							if p.Status != P2P_POINT_PAIRING {
								<-p.Chan
								continue
							}
							p0 = p
							continue
						default:
							continue
						}
					}
					select {
					case p.Chan <- true:
						if p0.Status != P2P_POINT_PAIRING {
							<-p0.Chan
							<-p.Chan
							p0 = nil
							goto bbreak
						}
						if p.Status != P2P_POINT_PAIRING {
							<-p.Chan
							continue
						}
						p0.Status = P2P_POINT_OFFER
						p.Status = P2P_POINT_ANSWER
						p0.Pair = p
						p.Pair = p0
						atomic.AddInt32(&Pairing_num, -2)
						<-p.Chan
						<-p0.Chan
						go p0.OnPair(p0, p)
						go p.OnPair(p, p0)
						goto bbreak
					default:
						continue
					}
				}
			}
			if p0 != nil {
				<-p0.Chan
			}

		bbreak:
			//time.Sleep(1 * time.Millisecond)
		}*/
}
