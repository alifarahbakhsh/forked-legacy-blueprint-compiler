package cache

import (
	"encoding/json"
	"sync"

	"github.com/bradfitz/gomemcache/memcache"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint-compiler/stdlib"
)

type MCConnection struct{}

func NewCacheConn() *MCConnection {
	return &MCConnection{}
}

type Memcached struct {
	client   *memcache.Client
	connPool *stdlib.ClientPool[*MCConnection]
}

func NewMemcachedClient(addr string, port string) *Memcached {
	conn_addr := addr + ":" + port
	client := memcache.New(conn_addr)
	client.MaxIdleConns = 60000
	pool := stdlib.NewClientPool[*MCConnection](1024, NewCacheConn)
	return &Memcached{client: client, connPool: pool}
}

func (m *Memcached) Put(key string, value interface{}) error {
	conn := m.connPool.Pop()
	defer m.connPool.Push(conn)
	marshaled_val, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return m.client.Set(&memcache.Item{Key: key, Value: marshaled_val})
}

func (m *Memcached) Get(key string, value interface{}) error {
	conn := m.connPool.Pop()
	defer m.connPool.Push(conn)
	it, err := m.client.Get(key)
	if err != nil {
		return err
	}
	return json.Unmarshal(it.Value, value)
}

func (m *Memcached) Incr(key string) (int64, error) {
	conn := m.connPool.Pop()
	defer m.connPool.Push(conn)
	val, err := m.client.Increment(key, 1)
	return int64(val), err
}

func (m *Memcached) Delete(key string) error {
	conn := m.connPool.Pop()
	defer m.connPool.Push(conn)
	return m.client.Delete(key)
}

func (m *Memcached) Mget(keys []string, values []interface{}) error {
	conn := m.connPool.Pop()
	defer m.connPool.Push(conn)
	val_map, err := m.client.GetMulti(keys)
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
			conn := m.connPool.Pop()
			defer wg.Done()
			defer m.connPool.Push(conn)
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
