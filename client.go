package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"searchinform/cache"
	"searchinform/provider"
)

// Duration - custom duration
type Duration struct {
	time.Duration
}

// MarshalJSON for json.Marshaler
func (d Duration) MarshalJSON() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, 32))
	buf.WriteByte('"')
	buf.WriteString(d.Duration.String())
	buf.WriteByte('"')
	return buf.Bytes(), nil
}

// UnmarshalJSON for json.Unmarshaler
func (d *Duration) UnmarshalJSON(data []byte) (err error) {
	data = data[1 : len(data)-1]
	d.Duration, err = time.ParseDuration(string(data))
	return
}

// Config - configuration format
type Config struct {
	Cache struct {
		TTL         Duration `json:"ttl"`
		NPartitions int      `json:"npartitions"`
	} `json:"cache"`

	Providers []provider.Provider `json:"providers"`

	HTTP struct {
		Port int `json:"port"`

		Timeout             Duration `json:"timeout"`
		DialTimeout         Duration `json:"dial_timeout"`
		KeepAliveTimeout    Duration `json:"keepalive_timeout"`
		TLSHandshakeTimeout Duration `json:"tls_handshake_timeout"`
	} `json:"http"`

	Log struct {
		IsDate         bool   `json:"is_date"`
		IsTime         bool   `json:"is_time"`
		IsMicroseconds bool   `json:"is_microseconds"`
		IsFile         bool   `json:"is_file"`
		Prefix         string `json:"prefix"`
	} `json:"log"`
}

// ParseConfig - parse config by file path
func ParseConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	conf := &Config{}
	if e := json.NewDecoder(file).Decode(conf); e != nil {
		return nil, e
	}
	return conf, nil
}

// LogFlags - converts config flags to log package format
func (conf *Config) LogFlags() (flags int) {
	fields := []struct {
		Flag int
		Need bool
	}{
		{Flag: log.Ldate, Need: conf.Log.IsDate},
		{Flag: log.Ltime, Need: conf.Log.IsTime},
		{Flag: log.Lmicroseconds, Need: conf.Log.IsMicroseconds},
		{Flag: log.Lshortfile, Need: conf.Log.IsFile},
	}
	for _, field := range fields {
		if field.Need {
			flags |= field.Flag
		}
	}
	return
}

// Client - main resolve client
type Client struct {
	cache     cache.Cache
	providers provider.Iterator
	client    http.Client
}

// NewClient - constuctor for Client struct
func NewClient(conf *Config) *Client {
	maxrate, providers := int64(0), conf.Providers
	for i := range providers {
		if rate := providers[i].MaxRate; maxrate < rate {
			maxrate = rate
		}
	}

	dialer := &net.Dialer{
		Timeout:   conf.HTTP.DialTimeout.Duration,
		KeepAlive: conf.HTTP.KeepAliveTimeout.Duration,
	}
	return &Client{
		cache:     *cache.NewCache(conf.Cache.NPartitions, conf.Cache.TTL.Duration),
		providers: *provider.NewIterator(conf.Providers),
		client: http.Client{
			Transport: &http.Transport{
				Dial:                dialer.Dial,
				DialContext:         dialer.DialContext,
				TLSHandshakeTimeout: conf.HTTP.TLSHandshakeTimeout.Duration,
				MaxIdleConnsPerHost: int(maxrate),
				IdleConnTimeout:     conf.HTTP.KeepAliveTimeout.Duration,
			},
			Timeout: conf.HTTP.Timeout.Duration,
		},
	}
}

func (c *Client) resolve(provider *provider.Provider, host string) (string, error) {
	url := fmt.Sprintf(provider.URLPattern, host)
	req, err := http.NewRequest(provider.Method, url, nil)
	if err != nil {
		return "", err
	}
	for key, value := range provider.Headers {
		req.Header.Set(key, value)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || 300 <= resp.StatusCode {
		return "", errors.New("Invalid status code:" + resp.Status)
	}

	return provider.ParseBody(resp.Body)
}

// Resolve returns country of this hosts
func (c *Client) Resolve(host string) (string, error) {
	if country, ok := c.cache.Get(host); ok {
		return country, nil
	}

	provider, err := c.providers.Next()
	if err != nil {
		return "", err
	}

	country, err := c.resolve(provider, host)
	if err != nil {
		return "", err
	}

	c.cache.Insert(host, country)

	return country, nil
}
