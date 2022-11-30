package redisdb

import (
	"edm/pkg/accs"
	"log"

	"github.com/go-redis/redis"
)

// NewReidsConnection is a constructor for the ObjectsInMemory type
func NewReidsConnection(arrsNames []string, addr string, passwd string, flushdb bool) *ObjectsInMemory {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: passwd,
		DB:       0,
	})
	objs := ObjectsInMemory{rdb: rdb}
	pong, err := rdb.Ping().Result()
	if err != nil {
		log.Println(accs.CurrentFunction(), pong, err)
	}
	if flushdb {
		objs.ClearAll()
	}
	objs.ResetBruteForceCounterImmediately()
	return &objs
}

// ObjectsInMemory points to redis connection
type ObjectsInMemory struct {
	rdb *redis.Client
}
