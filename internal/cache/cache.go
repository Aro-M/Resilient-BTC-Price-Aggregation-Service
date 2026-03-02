package cache

import (
	"sync"
	"time"
)

type ResponseData struct {
	Price       float64
	Currency    string
	SourcesUsed int
	LastUpdated time.Time
	Stale       bool
}

type PriceCache interface {
	Get(key string) (value ResponseData, ok bool)
	Set(key string, value ResponseData)
}

type priceCache struct {
	mu   sync.RWMutex
	data map[string]ResponseData
}

func New() PriceCache {
	return &priceCache{data: make(map[string]ResponseData)}
}

func (c *priceCache) Get(key string) (ResponseData, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.data[key]
	return v, ok
}

func (c *priceCache) Set(key string, value ResponseData) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
}
