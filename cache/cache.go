package cache

import (
	"hash/fnv"
	"sync/atomic"
	"time"
)

const (
	// TTL - time to live of cache entry
	TTL = 5 * time.Minute
)

// ValueType ...
type ValueType = string

// Entry - internal cache entry
type Entry struct {
	val      ValueType
	last     int64 // time of last data access (in UnixNano)
	deadline int64 // in UnixNano
}

// Cache - lock-free hash map
type Cache struct {
	partitions []list
}

// NewCache - create
func NewCache(npartitions int) *Cache {
	return &Cache{
		partitions: make([]list, npartitions),
	}
}

func (c *Cache) partition(key string) *list {
	hasher := fnv.New32a()
	hasher.Write([]byte(key))
	index := hasher.Sum32() % uint32(len(c.partitions))
	return &c.partitions[index]
}

// Get ...
func (c *Cache) Get(key string) (value ValueType, ok bool) {
	partition := c.partition(key)
	entry, okey := partition.Get(key)
	if !okey {
		return
	}

	now := time.Now().UnixNano()

	// check deadline
	if deadline := atomic.LoadInt64(&entry.deadline); now > deadline {
		partition.Delete(key)
		return
	}

	// update time of last data access
	for {
		last := atomic.LoadInt64(&entry.last)
		if last >= now || atomic.CompareAndSwapInt64(&entry.last, last, now) {
			break
		}
	}

	return entry.val, true
}

// Delete ...
func (c *Cache) Delete(key string) {
	partition := c.partition(key)
	partition.Delete(key)
}

// Insert ...
func (c *Cache) Insert(key string, value ValueType) {
	partition := c.partition(key)

	now := time.Now()
	entry := &Entry{
		val:      value,
		last:     now.UnixNano(),
		deadline: now.Add(TTL).UnixNano(),
	}
	partition.Insert(key, entry)
}
