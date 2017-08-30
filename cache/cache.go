package cache

import (
	"time"
)

const (
	// TTL - time to live of cache entry
	TTL = time.Minute

	npartitions = 32
)

// ValueType ...
type ValueType = uint8

// Entry - internal cache entry
type Entry struct {
	val      ValueType
	last     int64 // time of last data access (in UnixNano)
	deadline int64 // in UnixNano
}
