package cache

import (
	"strconv"
	"testing"
)

func TestCache(t *testing.T) {
	t.Parallel()

	entries := make([]ValueType, 32)
	for i := range entries {
		entries[i] = strconv.Itoa(i)
	}

	t.Run("insert+get", func(t *testing.T) {
		l := NewCache(4)

		// positive
		for _, entry := range entries {
			key := entry
			l.Insert(key, entry)
		}
		for _, entry := range entries {
			key := entry
			if value, ok := l.Get(key); !ok || value != entry {
				t.Fatalf("Get `%s` failed: expected: %v, but %v %v", key, entry, value, ok)
			}
		}

		// negative
		for _, key := range []string{"keys", "don't", "exist"} {
			if value, ok := l.Get(key); ok {
				t.Fatalf("Key `%s` doesn't exist, but returns %v %v", key, value, ok)
			}
		}
	})

	t.Run("delete", func(t *testing.T) {
		l := NewCache(8)
		l.Insert("zero", entries[0])
		l.Insert("one", entries[1])
		l.Insert("two", entries[2])

		const deleted = "one"
		l.Delete(deleted)
		if value, ok := l.Get(deleted); ok {
			t.Fatalf("Key `%s` has been deleted, but returns %v %v", deleted, value, ok)
		}

		if value, ok := l.Get("zero"); !ok || value != entries[0] {
			t.Fatalf("Get `zero` failed: expected: %v, but %v %v", entries[0], value, ok)
		}
		if value, ok := l.Get("two"); !ok || value != entries[2] {
			t.Fatalf("Get `two` failed: expected: %v, but %v %v", entries[2], value, ok)
		}
	})
}
