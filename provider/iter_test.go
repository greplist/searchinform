package provider

import "testing"

func TestIterPositive(t *testing.T) {
	t.Parallel()

	providers := []Provider{
		{Host: "host0", MaxRate: 2},
		{Host: "host1", MaxRate: 1},
		{Host: "host2", MaxRate: 2},
	}
	iter := NewIterator(providers)
	cases := []struct {
		Now  int64
		Host string
	}{
		{Now: 0, Host: "host0"},
		{Now: 59, Host: "host0"},
		{Now: 59, Host: "host1"},
		{Now: 60, Host: "host2"},
		{Now: 61, Host: "host2"},
		{Now: 119, Host: "host0"},
		{Now: 120, Host: "host0"},
	}
	for i, testCase := range cases {
		if provider, err := iter.next(testCase.Now); err != nil || provider.Host != testCase.Host {
			t.Fatalf("Iteration [%v]: must be host: `%v`, but actual: %v err: %v", i, testCase.Host, provider, err)
		}
	}
}

func TestIterNegative(t *testing.T) {
	t.Parallel()

	providers := []Provider{
		{Host: "host0", MaxRate: 1},
		{Host: "host1", MaxRate: 1},
	}
	iter := NewIterator(providers)
	cases := []struct {
		Now      int64
		Provider *Provider
		Err      error
	}{
		{Now: 0, Provider: &providers[0], Err: nil},
		{Now: 58, Provider: &providers[1], Err: nil},
		{Now: 59, Provider: nil, Err: ErrNotFound},
	}
	for i, testCase := range cases {
		provider, err := iter.next(testCase.Now)
		if testCase.Provider != nil && (provider == nil || provider.Host != testCase.Provider.Host) ||
			err != testCase.Err {
			t.Fatalf("Iteration [%v]: must be provoder: %v err: %v, but actual: %v err: %v",
				i, testCase.Provider, testCase.Err, provider, err)
		}
	}
}
