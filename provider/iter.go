package provider

import (
	"encoding/json"
	"errors"
	"io"
	"sync/atomic"
	"time"
)

var (
	// ErrNotFound ...
	ErrNotFound = errors.New("Not found provider, all providers are busy")
)

// Provider ...
type Provider struct {
	MaxRate    int64             `json:"max_rate"` // max number of requests per minute
	Name       string            `json:"name"`
	Method     string            `json:"method"`
	URLPattern string            `json:"pattern"`
	Scheme     []string          `json:"scheme"`
	Headers    map[string]string `json:"headers"`
}

// ParseBody returns country or error if body has invalid format
func (p *Provider) ParseBody(r io.Reader) (string, error) {
	var data map[string]interface{}
	if err := json.NewDecoder(r).Decode(&data); err != nil {
		return "", errors.New("Parse Body: json err: " + err.Error())
	}

	var value interface{} = data
	for _, field := range p.Scheme {
		m, ok := value.(map[string]interface{})
		if !ok {
			return "", errors.New("Invalid body format (field " + field + " type)")
		}
		value, ok = m[field]
		if !ok {
			return "", errors.New("Invalid body: field `" + field + "` not found")
		}
	}

	country, ok := value.(string)
	if !ok {
		return "", errors.New("Invalid body format (last field type)")
	}
	return country, nil
}

// ProvBlock - provider block for iterator
type ProvBlock struct {
	provider Provider
	rate     ReqRate
}

// Iterator - main struct
type Iterator struct {
	index  int32
	blocks []ProvBlock
}

// NewIterator - constructor for Iterator struct
func NewIterator(providers []Provider) *Iterator {
	blocks := make([]ProvBlock, 0, len(providers))
	for i := range providers {
		blocks = append(blocks, ProvBlock{provider: providers[i]})
	}
	return &Iterator{
		blocks: blocks,
	}
}

func (iter *Iterator) next(now int64) (provider *Provider, err error) {
	first, len := atomic.LoadInt32(&iter.index), int32(len(iter.blocks))

	index := first
	for {
		block := &iter.blocks[index]
		if rate := block.rate.rate(now); rate < block.provider.MaxRate {
			block.rate.observe(now)
			atomic.StoreInt32(&iter.index, index)
			return &block.provider, nil
		}

		if index = (index + 1) % len; index == first {
			break
		}
	}
	return nil, ErrNotFound
}

// Next - check request rate and returns next provider
func (iter *Iterator) Next() (provider *Provider, err error) {
	return iter.next(time.Now().Unix())
}
