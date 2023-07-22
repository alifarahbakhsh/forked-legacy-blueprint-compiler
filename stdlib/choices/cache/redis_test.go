package cache

import (
	"testing"
	"strconv"
	"time"
	"math/rand"
	"log"
	"sync"
)

func TestRedisPut(t *testing.T) {
	redis := NewRedisCacheClient("localhost", "6379")
	data := someData{ID: 5, Name: "Vaastav"}
	err := redis.Put("testData", data)
	if err != nil {
		t.Error(err)
	}
	var resultData someData
	err = redis.Get("testData", &resultData)
	if err != nil {
		t.Error(err)
	}
	if !equal(data, resultData) {
		t.Errorf("Incorrect data received from server: Expected: %v, Actual: %v", data, resultData)
	}
}

func TestRedisGet(t *testing.T) {
	redis := NewRedisCacheClient("localhost", "6379")
	var resultData someData
	err := redis.Get("testData", &resultData)
	if err != nil {
		t.Error(err)
	}
	if resultData.ID != 5 || resultData.Name != "Vaastav" {
		t.Errorf("Incorrect data received from server")
	}
}

func TestRedisIncr(t *testing.T) {
	redis := NewRedisCacheClient("localhost", "6379")
	err := redis.Put("intKey", 5)
	if err != nil {
		t.Error(err)
	}
	val, err := redis.Incr("intKey")
	if err != nil {
		t.Error(err)
	}
	if val != 6 {
		t.Errorf("Incorrect data received. Expected: 6, Actual %d", val)
	}
}

func TestRedisDelete(t *testing.T) {
	redis := NewRedisCacheClient("localhost", "6379")
	err := redis.Put("deleteKey", 6)
	if err != nil {
		t.Error(err)
	}
	var val int
	err = redis.Get("deleteKey", &val)
	if err != nil {
		t.Error(err)
	}
	if val != 6 {
		t.Errorf("Setup failed")
	}
	err = redis.Delete("deleteKey")
	if err != nil {
		t.Error(err)
	}
	var newval int
	err = redis.Get("deleteKey", &newval)
	if err == nil {
		t.Errorf("redis Cache miss didn't throw an error")
	}
	if newval != 0 {
		t.Errorf("Delete followed by a Get returned non-zero value")
	}
}

func TestRedisMget(t *testing.T) {
	var val1 someData
	var val2 int

	keys := []string{"testData", "intKey"}
	vals := []interface{}{&val1, &val2}
	redis := NewRedisCacheClient("localhost", "6379")
	err := redis.Mget(keys, vals)
	if err != nil {
		t.Error(err)
	}
	if val2 != 6 {
		t.Errorf("Incorrect value received from server. Expected: 6, Actual: %d", val2)
	}
	if val1.ID != 5 || val1.Name != "Vaastav" {
		t.Errorf("Incorrect value received from server. Expected: {5 Vaastav}, Actual: %v", val1)
	}
}

func TestRedisMset(t *testing.T) {
	redis := NewRedisCacheClient("localhost", "6379")
	keys := []string{"newKey", "testData", "intKey"}
	testData := someData{ID: 7, Name: "NotVaastav"}
	new_vals := []interface{}{6, testData, 5}

	err := redis.Mset(keys, new_vals)
	if err != nil {
		t.Error(err)
	}

	var val0 int
	var val1 someData
	var val2 int

	vals := []interface{}{&val0, &val1, &val2}
	err = redis.Mget(keys, vals)
	if err != nil {
		t.Error(err)
	}
	if val0 != 6 {
		t.Errorf("Incorrect value received from server. Expected: 6, Actual: %d", val0)
	}
	if val2 != 5 {
		t.Errorf("Incorrect value received from server. Expected: 5, Actual: %d", val2)
	}
	if val1.ID != 7 || val1.Name != "NotVaastav" {
		t.Errorf("Incorrect value received from server. Expected: {7 NotVaastav}, Actual: %v", val1)
	}
}

func TestRedisPerformance(t *testing.T) {
	redis := NewRedisCacheClient("localhost", "6379")
	keys := []string{}
	vals := []interface{}{}
	for i := 0; i < 100; i++ {
		keys = append(keys, strconv.FormatInt(int64(i),10))
		vals = append(vals, i)
	}
	err := redis.Mset(keys, vals)
	if err != nil {
		t.Error(err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 10000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var val int
			key_int := rand.Intn(100)
			key := strconv.FormatInt(int64(key_int), 10)
			start := time.Now()
			err := redis.Get(key,&val)
			if err != nil {
				t.Error(err)
			}
			duration := time.Now().Sub(start)
			log.Println(duration, val)
		}()
	}
	wg.Wait()
}