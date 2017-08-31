package cache

import "testing"

func TestList(t *testing.T) {
	t.Parallel()

	entries := make([]*Entry, 8)
	for i := range entries {
		entries[i] = &Entry{}
	}

	t.Run("insert+get", func(t *testing.T) {
		l := &list{}

		// positive
		cases := []struct {
			Key   string
			Value listValue
		}{
			{Key: "zero", Value: entries[0]},
			{Key: "one", Value: entries[1]},
			{Key: "two", Value: entries[2]},
		}
		for _, testCase := range cases {
			l.Insert(testCase.Key, testCase.Value)
		}
		for _, testCase := range cases {
			if value, ok := l.Get(testCase.Key); !ok || value != testCase.Value {
				t.Fatalf("Get `%s` failed: expected: %v, but %v %v", testCase.Key, testCase.Value, value, ok)
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
		l := &list{}
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
