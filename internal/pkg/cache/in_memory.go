package cache

import (
	"sync"
	"time"
)

type Clock interface {
	Now() time.Time
}

type InMemoryCache struct {
	dataMap map[int]inMemoryValue
	lock    *sync.Mutex
	clock   Clock
}

type inMemoryValue struct {
	SetTime    int64
	Expiration int64
}

//func InitInMemoryCache(clock Clock) *InMemoryCache {
//	return &InMemoryCache{
//		dataMap: make(map[int]inMemoryValue, 0),
//		lock:    &sync.Mutex{},
//		clock:   clock,
//	}
//}

func (c *InMemoryCache) Add(key int, expiration int64) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.dataMap[key] = inMemoryValue{
		SetTime:    c.clock.Now().Unix(),
		Expiration: expiration,
	}
	return nil
}

func (c *InMemoryCache) Get(key int) (bool, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	value, ok := c.dataMap[key]
	if ok && c.clock.Now().Unix()-value.SetTime > value.Expiration {
		return false, nil
	}
	return ok, nil
}

func (c *InMemoryCache) Delete(key int) {
	c.lock.Lock()
	defer c.lock.Unlock()
	delete(c.dataMap, key)
}
