package cache

import (
	"sync/atomic"
	"unsafe"
)

// ListValue ...
type listValue = Entry

// node of hlist
type node struct {
	next unsafe.Pointer // real type is *node

	key   string
	value listValue
}

func (n *node) Next() *node {
	return (*node)(atomic.LoadPointer(&n.next))
}

// list - lock-free list for Hash map
type list struct {
	head unsafe.Pointer // real type is *node
}

func (l *list) Head() *node {
	return (*node)(atomic.LoadPointer(&l.head))
}

func (l *list) Insert(key string, value listValue) {
	ptr := &node{
		key:   key,
		value: value,
	}
	new := unsafe.Pointer(ptr)
	for {
		old := atomic.LoadPointer(&l.head)
		ptr.next = old
		if atomic.CompareAndSwapPointer(&l.head, old, new) {
			break
		}
	}
}

func (l *list) Get(key string) (value listValue, ok bool) {
	for this := l.Head(); this != nil; this = this.Next() {
		if this.key == key {
			return this.value, true
		}
	}
	return
}
