package gate

// 单机,本地的用户路由

import (
	"sync"

	"github.com/ipiao/meim/log"
)

type Router struct {
	mu      sync.RWMutex
	clients map[int64]ClientSet // 这个必须有userId
}

func NewRouter() *Router {
	route := new(Router)
	route.clients = make(map[int64]ClientSet)
	return route
}

// uid 已经设置的情况下才可调用
//
func (route *Router) AddClient(client *Client) {
	route.mu.Lock()
	defer route.mu.Unlock()
	set, ok := route.clients[client.UID()]
	if !ok {
		set = NewClientSet()
		route.clients[client.UID()] = set
	}
	set.Add(client)
}

func (route *Router) RemoveClient(client *Client) bool {
	route.mu.Lock()
	defer route.mu.Unlock()
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

func (route *Router) FindClientSet(uid int64) ClientSet {
	route.mu.RLock()
	defer route.mu.RUnlock()

	set, ok := route.clients[uid]
	if ok {
		return set.Clone()
	} else {
		return nil
	}
}

func (route *Router) IsOnline(uid int64) bool {
	route.mu.RLock()
	defer route.mu.RUnlock()

	set, ok := route.clients[uid]
	if ok {
		return len(set) > 0
	}
	return false
}

type ClientSet map[*Client]struct{}

func NewClientSet() ClientSet {
	return make(map[*Client]struct{})
}

func (set ClientSet) Add(c *Client) {
	set[c] = struct{}{}
}

func (set ClientSet) IsMember(c *Client) bool {
	if _, ok := set[c]; ok {
		return true
	}
	return false
}

func (set ClientSet) Remove(c *Client) {
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
	n := make(map[*Client]struct{})
	for k, v := range set {
		n[k] = v
	}
	return n
}
