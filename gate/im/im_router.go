package gate

import (
	"sync"

	"github.com/ipiao/meim/gate"
	"github.com/ipiao/meim/log"
)

// user router
type Route struct {
	mutex   sync.Mutex
	clients map[int64]ClientSet
}

func NewRoute() *Route {
	route := new(Route)
	route.clients = make(map[int64]ClientSet)
	return route
}

// uid 已经设置的情况下才可调用
//
func (route *Route) AddClient(client *gate.Client) {
	route.mutex.Lock()
	defer route.mutex.Unlock()
	set, ok := route.clients[client.UID()]
	if !ok {
		set = NewClientSet()
		route.clients[client.UID()] = set
	}
	set.Add(client)
}

func (route *Route) RemoveClient(client *gate.Client) bool {
	route.mutex.Lock()
	defer route.mutex.Unlock()
	if set, ok := route.clients[client.UID()]; ok {
		set.Remove(client)
		if set.Count() == 0 {
			delete(route.clients, client.UID())
		}
		return true
	}
	log.Infof("client non exists, %s", client.Log())
	return false
}

func (route *Route) FindClientSet(uid int64) ClientSet {
	route.mutex.Lock()
	defer route.mutex.Unlock()

	set, ok := route.clients[uid]
	if ok {
		return set.Clone()
	} else {
		return nil
	}
}

func (route *Route) IsOnline(uid int64) bool {
	route.mutex.Lock()
	defer route.mutex.Unlock()

	set, ok := route.clients[uid]
	if ok {
		return len(set) > 0
	}
	return false
}

type ClientSet map[*gate.Client]struct{}

func NewClientSet() ClientSet {
	return make(map[*gate.Client]struct{})
}

func (set ClientSet) Add(c *gate.Client) {
	set[c] = struct{}{}
}

func (set ClientSet) IsMember(c *gate.Client) bool {
	if _, ok := set[c]; ok {
		return true
	}
	return false
}

func (set ClientSet) Remove(c *gate.Client) {
	if _, ok := set[c]; !ok {
		return
	}
	delete(set, c)
}

func (set ClientSet) Count() int {
	return len(set)
}

// 只是浅复制
func (set ClientSet) Clone() ClientSet {
	n := make(map[*gate.Client]struct{})
	for k, v := range set {
		n[k] = v
	}
	return n
}
