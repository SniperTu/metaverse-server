package conf

import (
	"fmt"
	"log"

	"github.com/go-redis/redis"
)

var RedisClient *redis.Client

func init() {
	log.Println(fmt.Sprintf("Try to connect to Redis host: %s", Conf.Db.Redis.Host))
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     Conf.Db.Redis.Host,
		Password: Conf.Db.Redis.Pwd,
		DB:       Conf.Db.Redis.Database,
	})
	if _, err := RedisClient.Ping().Result(); err != nil {
		log.Println("Redis初始化错误", err)
		return
	}
}
