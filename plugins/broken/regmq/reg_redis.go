package regmq

import (
	"time"

	"github.com/go-redis/redis"
	"github.com/ipiao/meim/log"
)

type KeyFunc func(int64) string

// Redis注册
type RedisRegistry struct {
	client         *redis.Client
	keyFn          KeyFunc
	internalClient bool
}

func NewRedisRegistry(client *redis.Client, keyFn KeyFunc) *RedisRegistry {
	return &RedisRegistry{
		client: client,
		keyFn:  keyFn,
	}
}

func NewRedisRegistry2(address, password string, db int, keyFn KeyFunc) *RedisRegistry {
	return &RedisRegistry{
		client:         newRedisClient(address, password, db),
		keyFn:          keyFn,
		internalClient: true,
	}
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

// TODO 允许web端和手机端同时登录,通过禁用web端某些功能保证业务一致性
// 暂时只是允许单点登录
func (r *RedisRegistry) Register(uid int64, node int) {
	key := r.keyFn(uid)
	err := r.client.SetNX(key, node, 0).Err()
	if err != nil {
		log.Errorf("SetNX Error: %s", err)
	}
}

func (r *RedisRegistry) DeRegister(uid int64) {
	key := r.keyFn(uid)
	err := r.client.Del(key).Err()
	if err != nil && err != redis.Nil {
		log.Errorf("Del Error: %s", err)
	}
}

func (r *RedisRegistry) GetUserNode(uid int64) int {
	key := r.keyFn(uid)
	val, err := r.client.Get(key).Int()
	if err != nil && err != redis.Nil {
		log.Errorf("Get Error: %s", err)
	}
	return val
}

func (r *RedisRegistry) Close() {
	if r.internalClient {
		r.client.Close()
	}
}
