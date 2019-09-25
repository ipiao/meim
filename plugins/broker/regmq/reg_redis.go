package regmq

import (
	"fmt"
	"time"

	"github.com/go-redis/redis"
	"github.com/ipiao/meim/log"
)

type KeyFunc func(int64) string

// Redis注册
type RedisRegistry struct {
	client         *redis.Client
	key            string
	internalClient bool
}

func NewRedisRegistry(client *redis.Client, key string) *RedisRegistry {
	return &RedisRegistry{
		client: client,
		key:    key,
	}
}

func NewRedisRegistry2(address, password string, db int, key string) *RedisRegistry {
	return &RedisRegistry{
		client:         newRedisClient(address, password, db),
		key:            key,
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
	err := r.client.HSet(r.key, fmt.Sprint(uid), node).Err()
	if err != nil {
		log.Errorf("HSet Error: %s", err)
	}
}

func (r *RedisRegistry) DeRegister(uid int64) {
	err := r.client.HDel(r.key, fmt.Sprint(uid)).Err()
	if err != nil && err != redis.Nil {
		log.Errorf("HDel Error: %s", err)
	}
}

func (r *RedisRegistry) GetUserNode(uid int64) int {
	val, err := r.client.HGet(r.key, fmt.Sprint(uid)).Int()
	if err != nil && err != redis.Nil {
		log.Errorf("HGet Error: %s", err)
	}
	return val
}

func (r *RedisRegistry) Close() {
	if r.internalClient {
		r.client.Close()
	}
}
