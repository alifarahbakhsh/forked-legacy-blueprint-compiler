package cache

import (
	"encoding/json"
	"sync"

	"github.com/bradfitz/gomemcache/memcache"
)

type Memcached struct {
	Client *memcache.Client
}

func NewMemcachedClient(addr string, port string) *Memcached {
	conn_addr := addr + ":" + port
	client := memcache.New(conn_addr)
	client.MaxIdleConns = 1000
	return &Memcached{Client: client}
}

func (m *Memcached) Put(key string, value interface{}) error {
	marshaled_val, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return m.Client.Set(&memcache.Item{Key: key, Value: marshaled_val})
}

func (m *Memcached) Get(key string, value interface{}) error {
	it, err := m.Client.Get(key)
	if err != nil {
		return err
	}
	return json.Unmarshal(it.Value, value)
}

func (m *Memcached) Incr(key string) (int64, error) {
	val, err := m.Client.Increment(key, 1)
	return int64(val), err
}

func (m *Memcached) Delete(key string) error {
	return m.Client.Delete(key)
}

func (m *Memcached) Mget(keys []string, values []interface{}) error {
	val_map, err := m.Client.GetMulti(keys)
	if err != nil {
		return err
	}
	for idx, key := range keys {
		if val, ok := val_map[key]; ok {
			err := json.Unmarshal(val.Value, values[idx])
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *Memcached) Mset(keys []string, values []interface{}) error {
	var wg sync.WaitGroup
	wg.Add(len(keys))
	err_chan := make(chan error, len(keys))
	for idx, key := range keys {
		go func(key string, val interface{}) {
			defer wg.Done()
			err_chan <- m.Put(key, val)
		}(key, values[idx])
	}
	wg.Wait()
	close(err_chan)
	for err := range err_chan {
		if err != nil {
			return err
		}
	}
	return nil
}
