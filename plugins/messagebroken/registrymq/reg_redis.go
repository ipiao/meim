package registrymq

import (
	"time"

	"github.com/go-redis/redis"
)

// Redis注册
type RedisRegistry struct {
	client *redis.Client
	keyFn  func(int64) string
}

func newRedisClient(address, password string, db int) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:        address,
		Password:    password,
		DB:          db,
		DialTimeout: time.Second * 2,
		MaxConnAge:  time.Minute * 30,

		OnConnect: func(conn *redis.Conn) error {
			return conn.Ping().Err()
		},
	})
	return client
}

func (r *RedisRegistry) Register(uid int64, node int) {
	key := r.keyFn(uid)
	r.client.SAdd(key, node)
}

func (r *RedisRegistry) DeRegister(uid int64, node int) {
	key := r.keyFn(uid)
	r.client.SRem(key, node)
}
