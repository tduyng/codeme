package stats

import (
	"sync"
	"time"
)

type StatsCache struct {
	mu        sync.RWMutex
	stats     *APIStats
	generated time.Time
	ttl       time.Duration
}

func NewStatsCache(ttl time.Duration) *StatsCache {
	if ttl == 0 {
		ttl = 30 * time.Second
	}
	return &StatsCache{ttl: ttl}
}

func (c *StatsCache) Get(opts APIOptions) (*APIStats, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.stats == nil {
		return nil, false
	}

	if time.Since(c.generated) > c.ttl {
		return nil, false
	}

	return c.stats, true
}

func (c *StatsCache) Set(stats *APIStats) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.stats = stats
	c.generated = time.Now()
}

func (c *StatsCache) Invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.stats = nil
}
