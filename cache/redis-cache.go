package cache

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"example.com/go-demo/models"
	"github.com/go-redis/redis/v7"
)

type redisCache struct {
	host    string
	db      int
	expires time.Duration
}

func NewRedisCache(host string, db int, exp time.Duration) CoviddataCache {
	return &redisCache{
		host:    host,
		db:      db,
		expires: exp,
	}
}

func (cache *redisCache) getClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     cache.host,
		Password: string(os.Getenv("REDIS_DB_PASSWORD")),
		DB:       cache.db,
	})
}

func (cache *redisCache) Set(key string, coviddata *models.Coviddata) {
	client := cache.getClient()

	// serialize Post object to JSON
	json, err := json.Marshal(coviddata)
	if err != nil {
		panic(err)
	}

	log.Println(" object cached: " + string(json) + " cache expiring time : " + (cache.expires * time.Minute).String() + " cached using key : " + key)
	log.Println(client.ClientList())
	err = client.Set(key, json, cache.expires*time.Minute).Err()
	if err != nil {
		panic(err)
	}
}

func (cache *redisCache) Get(key string) *models.Coviddata {
	client := cache.getClient()

	val, err := client.Get(key).Result()
	if err != nil {
		return nil
	}
	log.Println(client.ClientList())

	coviddata := models.Coviddata{}
	err = json.Unmarshal([]byte(val), &coviddata)
	if err != nil {
		panic(err)
	}

	return &coviddata
}
