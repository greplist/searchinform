package cache

import (
	"sync/atomic"
	"unsafe"
)

// ListValue ...
type listValue = *Entry

// node of list
type node struct {
	next unsafe.Pointer // real type is *node

	key   string
	value listValue
}

func loadNode(ptr *unsafe.Pointer) *node {
	return (*node)(atomic.LoadPointer(ptr))
}

func (n *node) Next() *node {
	return loadNode(&n.next)
}

// list - lock-free list for Hash map
type list struct {
	head unsafe.Pointer // real type is *node
}

func (l *list) Head() *node {
	return (*node)(atomic.LoadPointer(&l.head))
}

func find(start *node, key string) *node {
	for this := start; this != nil; this = this.Next() {
		if this.key == key {
			return this
		}
	}
	return nil
}

func (l *list) tryRemove(entry *node) bool {
	indirect := &l.head

	for this := loadNode(indirect); this != entry; this = loadNode(indirect) {
		// already deleted by another thread
		if this == nil {
			return false
		}
		indirect = &this.next
	}

	old, new := unsafe.Pointer(entry), atomic.LoadPointer(&entry.next)
	return atomic.CompareAndSwapPointer(indirect, old, new)
}

func (l *list) Get(key string) (value listValue, ok bool) {
	if entry := find(l.Head(), key); entry != nil {
		return entry.value, true
	}
	return
}

func (l *list) Delete(key string) {
	for entry := find(l.Head(), key); entry != nil; entry = find(l.Head(), key) {
		if ok := l.tryRemove(entry); ok {
			break
		}
	}
}

func (l *list) Insert(key string, value listValue) {
	// push to the top
	ptr := &node{
		key:   key,
		value: value,
	}
	new := unsafe.Pointer(ptr)
	for {
		old := atomic.LoadPointer(&l.head)
		if ptr.next = old; atomic.CompareAndSwapPointer(&l.head, old, new) {
			break
		}
	}

	// remove all old nodes before new one
	for {
		entry := find(ptr.Next(), key)
		if entry == nil {
			break
		}
		l.tryRemove(entry)
	}
}
