package cache

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
)

type RedisCache struct {
	client *redis.Client
}

func NewRedisCacheClient(addr string, port string) *RedisCache {
	conn_addr := addr + ":" + port
	client := redis.NewClient(&redis.Options{
		Addr: conn_addr,
		Password: "",
		DB: 0,
	})
	return &RedisCache{client: client}
}

func (r *RedisCache) Put(key string, value interface{}) error {
	ctx := context.Background()
	val, err := json.Marshal(value)
	if err != nil {
		return err
	}
	val_str := string(val)
	return r.client.Set(ctx, key, val_str, 0).Err()
}

func (r *RedisCache) Get(key string, value interface{}) error {
	ctx := context.Background()
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), value)
}

func (r *RedisCache) Incr(key string) (int64, error) {
	ctx := context.Background()
	return r.client.Incr(ctx, key).Result()
}

func (r *RedisCache) Delete(key string) error {
	ctx := context.Background()
	return r.client.Del(ctx, key).Err()
}

func (r *RedisCache) Mget(keys []string, values []interface{}) error {
	ctx := context.Background()
	result, err := r.client.MGet(ctx, keys...).Result()
	if err != nil {
		return err
	}
	for idx, res := range result {
		err := json.Unmarshal([]byte(res.(string)), values[idx])
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *RedisCache) Mset(keys []string, values []interface{}) error {
	ctx := context.Background()
	kv_map := make(map[string]string)
	for idx, key := range keys {
		val, err := json.Marshal(values[idx])
		if err != nil {
			return err
		}
		kv_map[key] = string(val)
	}
	return r.client.MSet(ctx, kv_map).Err()
}