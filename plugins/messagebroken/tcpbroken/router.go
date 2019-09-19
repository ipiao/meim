package tcpbroken

import (
	"sync"

	"github.com/ipiao/meim/util"
)

// 对应每个comect的Route,记录每个comect的用户信息
type Route struct {
	mutex sync.Mutex
	uids  util.IntSet
	// roomIds util.IntSet
}

func NewRoute() *Route {
	r := new(Route)
	r.uids = util.NewIntSet()
	// r.roomIds = util.NewIntSet()

	return r
}

func (route *Route) ContainUserID(uid int64) bool {
	route.mutex.Lock()
	defer route.mutex.Unlock()
	_, ok := route.uids[uid]

	return ok
}

func (route *Route) IsUserOnline(uid int64) bool {
	route.mutex.Lock()
	defer route.mutex.Unlock()

	return route.uids.IsMember(uid)
}

func (route *Route) AddUserID(uid int64) {
	route.mutex.Lock()
	defer route.mutex.Unlock()

	route.uids.Add(uid)
}

func (route *Route) RemoveUserID(uid int64) {
	route.mutex.Lock()
	defer route.mutex.Unlock()

	delete(route.uids, uid)
}

func (route *Route) GetUserIDs() util.IntSet {
	route.mutex.Lock()
	defer route.mutex.Unlock()

	uids := util.NewIntSet()
	for uid := range route.uids {
		uids.Add(uid)
	}
	return uids
}

// func (route *Route) ContainRoomID(room_id int64) bool {
// 	route.mutex.Lock()
// 	defer route.mutex.Unlock()

// 	return route.roomIds.IsMember(room_id)
// }

// func (route *Route) AddRoomID(room_id int64) {
// 	route.mutex.Lock()
// 	defer route.mutex.Unlock()

// 	route.roomIds.Add(room_id)
// }

// func (route *Route) RemoveRoomID(room_id int64) {
// 	route.mutex.Lock()
// 	defer route.mutex.Unlock()

// 	route.roomIds.Remove(room_id)
// }

//=================client===
