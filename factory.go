package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/searchinform/cache"
	"github.com/searchinform/provider"
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

// Factory - main abstract factory for Cache & HTTPClient
type Factory struct {
	Config *Config
}

// NewFactory - constructor for Factory struct
func NewFactory(conf *Config) *Factory {
	return &Factory{Config: conf}
}

// NewLogger returns logger with correct settings
func (f *Factory) NewLogger() *log.Logger {
	conf := f.Config.Log

	var flags int
	fields := []struct {
		Flag int
		Need bool
	}{
		{Flag: log.Ldate, Need: conf.IsDate},
		{Flag: log.Ltime, Need: conf.IsTime},
		{Flag: log.Lmicroseconds, Need: conf.IsMicroseconds},
		{Flag: log.Lshortfile, Need: conf.IsFile},
	}
	for _, field := range fields {
		if field.Need {
			flags |= field.Flag
		}
	}
	return log.New(os.Stdout, conf.Prefix, flags)
}

// NewCache returns cache with correct settings
func (f *Factory) NewCache() *cache.Cache {
	conf := &f.Config.Cache
	return cache.NewCache(conf.NPartitions, conf.TTL.Duration)
}

// NewProviders returns provider list with correct settings
func (f *Factory) NewProviders() *provider.Iterator {
	return provider.NewIterator(f.Config.Providers)
}

// NewDefaultHTTPClient returns http.Client with correct settings
func (f *Factory) NewDefaultHTTPClient() *http.Client {
	maxrate, providers := int64(0), f.Config.Providers
	for i := range providers {
		if rate := providers[i].MaxRate; maxrate < rate {
			maxrate = rate
		}
	}

	conf := &f.Config.HTTP
	dialer := &net.Dialer{
		Timeout:   conf.DialTimeout.Duration,
		KeepAlive: conf.KeepAliveTimeout.Duration,
	}
	return &http.Client{
		Transport: &http.Transport{
			Dial:                dialer.Dial,
			DialContext:         dialer.DialContext,
			TLSHandshakeTimeout: conf.TLSHandshakeTimeout.Duration,
			MaxIdleConnsPerHost: int(maxrate),
			IdleConnTimeout:     conf.KeepAliveTimeout.Duration,
		},
		Timeout: conf.Timeout.Duration,
	}
}

// NewHTTPClient returns custom HTTP Client with correct settings
func (f *Factory) NewHTTPClient() *HTTPClient {
	return NewHTTPClient(f.NewDefaultHTTPClient(), f.NewProviders())
}

// NewController returns Controller with correct settings
func (f *Factory) NewController() *Controller {
	return &Controller{
		cache:     *f.NewCache(),
		providers: *f.NewProviders(),
		client:    *f.NewHTTPClient(),
		logger:    *f.NewLogger(),
	}
}
