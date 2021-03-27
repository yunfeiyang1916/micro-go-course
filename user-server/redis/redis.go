package redis

import (
	"fmt"
	"github.com/go-redsync/redsync"
	"github.com/gomodule/redigo/redis"
	"time"
)

// 缓存池
var pool *redis.Pool

// 分布式锁
var redisLock *redsync.Redsync

// 初始化redis
func InitRedis(host, port, password string) error {
	pool = &redis.Pool{
		MaxIdle:     20,
		IdleTimeout: 240 * time.Second,
		MaxActive:   50,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", fmt.Sprintf("%s:%s", host, port))
			if err != nil {
				return nil, err
			}
			if password != "" {
				if _, err := c.Do("AUTH", password); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
	redisLock = redsync.New([]redsync.Pool{pool})
	return nil
}

// 获取redis连接
func GetRedisConn() (redis.Conn, error) {
	conn := pool.Get()
	return conn, conn.Err()
}

// 获取分布式锁
func GetRedisLock(key string, expireTime time.Duration) *redsync.Mutex {
	return redisLock.NewMutex(key, redsync.SetExpiry(expireTime))
}
