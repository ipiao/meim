package meim

// 单机,本地的用户路由

import (
	"sync"

	"github.com/ipiao/meim/log"
)

// Router 进行用户客户端管理/查找的路由服务
// example
type Router struct {
	mu      sync.RWMutex        //
	clients map[int64]ClientSet // 这个必须有userId,和server.clients的生存周期有所不同
}

func NewRouter() *Router {
	route := new(Router)
	route.clients = make(map[int64]ClientSet)
	return route
}

// uid 已经设置的情况下才可调用
// 不允许uid为0
func (route *Router) AddClient(client *Client) {
	if client.UID == 0 {
		log.Warnf("router add invalid client %s", client.Log())
		return
	}
	route.mu.Lock()
	defer route.mu.Unlock()
	set, ok := route.clients[client.UID]
	if !ok {
		set = NewClientSet()
		route.clients[client.UID] = set
	}
	set.Add(client)
}

func (route *Router) RemoveClient(client *Client) bool {
	route.mu.Lock()
	defer route.mu.Unlock()
	if set, ok := route.clients[client.UID]; ok {
		set.Remove(client)
		if set.Count() == 0 {
			delete(route.clients, client.UID)
		}
		return true
	}
	log.Infof("router client non exists, %s", client.Log())
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

// FindClient 只查找一个在线client
func (route *Router) FindClient(uid int64) *Client {
	route.mu.RLock()
	defer route.mu.RUnlock()

	set, ok := route.clients[uid]
	if ok {
		for c := range set {
			if !c.closed.Load() {
				return c
			}
		}
	}
	return nil
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
