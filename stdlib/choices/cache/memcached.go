package cache

import (
	"encoding/json"
	"sync"

	"github.com/bradfitz/gomemcache/memcache"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint-compiler/stdlib"
)

type MCConnection struct{
	Client *memcache.Client
}

func NewCacheConn(addr string, port string) *MCConnection {
	conn_addr := addr + ":" + port
	client := memcache.New(conn_addr)
	client.MaxIdleConns = 60000
	return &MCConnection{Client : client}
}

type Memcached struct {
	connPool *stdlib.ClientPool[*MCConnection]
}

func NewMemcachedClient(addr string, port string) *Memcached {
	NewConn := func() *MCConnection {
		return NewCacheConn(addr, port)
	}
	pool := stdlib.NewClientPool[*MCConnection](1024, NewConn)
	return &Memcached{connPool: pool}
}

func (m *Memcached) Put(key string, value interface{}) error {
	conn := m.connPool.Pop()
	defer m.connPool.Push(conn)
	marshaled_val, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return conn.Client.Set(&memcache.Item{Key: key, Value: marshaled_val})
}

func (m *Memcached) Get(key string, value interface{}) error {
	conn := m.connPool.Pop()
	defer m.connPool.Push(conn)
	it, err := conn.Client.Get(key)
	if err != nil {
		return err
	}
	return json.Unmarshal(it.Value, value)
}

func (m *Memcached) Incr(key string) (int64, error) {
	conn := m.connPool.Pop()
	defer m.connPool.Push(conn)
	val, err := conn.Client.Increment(key, 1)
	return int64(val), err
}

func (m *Memcached) Delete(key string) error {
	conn := m.connPool.Pop()
	defer m.connPool.Push(conn)
	return conn.Client.Delete(key)
}

func (m *Memcached) Mget(keys []string, values []interface{}) error {
	conn := m.connPool.Pop()
	defer m.connPool.Push(conn)
	val_map, err := conn.Client.GetMulti(keys)
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
