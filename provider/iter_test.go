package provider

import (
	"strings"
	"testing"
)

func TestProviderBodyParsePositive(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Body    string
		Scheme  []string
		Country string
	}{
		{Body: `{"a":"Belarus"}`, Scheme: []string{"a"}, Country: "Belarus"},
		{Body: `{"a":{"b":"Belarus"}}`, Scheme: []string{"a", "b"}, Country: "Belarus"},
		{Body: `{"a":{"b":{"c":"Belarus"}}}`, Scheme: []string{"a", "b", "c"}, Country: "Belarus"},
	}
	for _, testCase := range cases {
		provider := &Provider{Scheme: testCase.Scheme}
		body := strings.NewReader(testCase.Body)
		if country, err := provider.ParseBody(body); err != nil || country != testCase.Country {
			t.Fatalf("Invalid country or err: expected `%s`, but country : `%s` err : %v", testCase.Country, country, err)
		}
	}
}

func TestProviderBodyParseNegative(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Body   string
		Scheme []string
		Err    string
	}{
		{Body: `{"a":"Belarus"}`, Scheme: []string{"i"}, Err: "not found field i"},
		{Body: `{"a":{"b":{"c":"Belarus"}}}`, Scheme: []string{"i", "b", "c"}, Err: "not found field i"},
		{Body: `{"a":{"b":{"c":"Belarus"}}}`, Scheme: []string{"a", "i", "c"}, Err: "not found field i"},
		{Body: `{"a":{"b":{"c":"Belarus"}}}`, Scheme: []string{"a", "b", "i"}, Err: "not found field i"},
		{Body: `{"a": 1}`, Scheme: []string{"a"}, Err: "invalid type field a"},
		{Body: `{"a":{"b":"Belarus"}}`, Scheme: []string{"a", "b", "c"}, Err: "invalid type field b"},
		{Body: `{"a":{"b":{"c":1}}}`, Scheme: []string{"a", "b", "c"}, Err: "not found field c"},
	}
	for _, testCase := range cases {
		provider := &Provider{Scheme: testCase.Scheme}
		body := strings.NewReader(testCase.Body)
		if _, err := provider.ParseBody(body); err == nil {
			t.Fatalf("No error, but must be %s", testCase.Err)
		}
	}
}

func TestIterPositive(t *testing.T) {
	t.Parallel()

	providers := []Provider{
		{URLPattern: "host0", MaxRate: 2},
		{URLPattern: "host1", MaxRate: 1},
		{URLPattern: "host2", MaxRate: 2},
	}
	iter := NewIterator(providers)
	cases := []struct {
		Now        int64
		URLPattern string
	}{
		{Now: 0, URLPattern: "host0"},
		{Now: 59, URLPattern: "host0"},
		{Now: 59, URLPattern: "host1"},
		{Now: 60, URLPattern: "host2"},
		{Now: 61, URLPattern: "host2"},
		{Now: 119, URLPattern: "host0"},
		{Now: 120, URLPattern: "host0"},
	}
	for i, testCase := range cases {
		if provider, err := iter.next(testCase.Now); err != nil || provider.URLPattern != testCase.URLPattern {
			t.Fatalf("Iteration [%v]: must be host: `%v`, but actual: %v err: %v", i, testCase.URLPattern, provider, err)
		}
	}
}

func TestIterNegative(t *testing.T) {
	t.Parallel()

	providers := []Provider{
		{URLPattern: "host0", MaxRate: 1},
		{URLPattern: "host1", MaxRate: 1},
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
		if testCase.Provider != nil && (provider == nil || provider.URLPattern != testCase.Provider.URLPattern) ||
			err != testCase.Err {
			t.Fatalf("Iteration [%v]: must be provoder: %v err: %v, but actual: %v err: %v",
				i, testCase.Provider, testCase.Err, provider, err)
		}
	}
}
