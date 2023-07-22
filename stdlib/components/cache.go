package components

type Cache interface {
	Put(key string, value interface{}) error
	// val is the pointer to which the value will be stored
	Get(key string, val interface{}) error
	Mset(keys []string, values []interface{}) error
	// values is the array of pointers to which the value will be stored 
	Mget(keys []string, values []interface{}) error
	Delete(key string) error
	Incr(key string) (int64, error)
}