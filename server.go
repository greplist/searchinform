package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"log"
	"net"
	"net/http"
	"strconv"

	"searchinform/cache"
	"searchinform/provider"
)

var (
	configPath string
)

func init() {
	flag.StringVar(&configPath, "c", "conf.json", "config filepath")
}

// Controller - main struct with all dependences
type Controller struct {
	cache     cache.Cache
	providers provider.Iterator
	client    HTTPClient
	logger    log.Logger
}

// Init run all background jobs
func (ctrl *Controller) Init() {
	go cache.Cleaner(context.Background(), &ctrl.cache)
}

func (ctrl *Controller) error(w http.ResponseWriter, msg string, code int) {
	ctrl.logger.Println(msg)
	http.Error(w, msg, code)
}

func lookup(host string) (addr string, err error) {
	addrs, err := net.LookupHost(host)
	if err != nil {
		return "", err
	}
	if len(addrs) == 0 {
		return "", errors.New("empty addrs list")
	}
	return addrs[0], nil
}

// resolve returns country of this host
func (ctrl *Controller) resolve(host string) (string, error) {
	addr, err := lookup(host)
	if err != nil {
		return "", errors.New("host lookup err : " + err.Error())
	}

	if country, ok := ctrl.cache.Get(addr); ok {
		return country, nil
	}

	provider, err := ctrl.providers.Next()
	if err != nil {
		return "", errors.New("providers iter err : " + err.Error())
	}

	country, err := ctrl.client.Resolve(provider, addr)
	if err != nil {
		return "", errors.New("http client err : " + err.Error())
	}

	ctrl.cache.Insert(addr, country)

	return country, nil
}

// CountryByIP ..
func (ctrl *Controller) CountryByIP(w http.ResponseWriter, r *http.Request) {
	host := r.FormValue("host")
	if host == "" {
		host = r.Host
	}

	country, err := ctrl.resolve(host)
	if err != nil {
		ctrl.error(w, "Resolve err: "+err.Error(), http.StatusInternalServerError)
		return
	}

	body := &struct {
		Host    string `json:"host"`
		Country string `json:"country"`
	}{Host: host, Country: country}

	json.NewEncoder(w).Encode(body)
}

func main() {
	flag.Parse()

	log.Println("Parsing config file", configPath)
	conf, err := ParseConfig(configPath)
	if err != nil {
		log.Fatalln("Parse err:", err)
	}

	ctrl := NewFactory(conf).NewController()
	ctrl.Init()

	router := http.NewServeMux()
	router.HandleFunc("/api/country", ctrl.CountryByIP)

	port := conf.HTTP.Port
	log.Printf("Server start on %v port...\n", port)
	if err := http.ListenAndServe(":"+strconv.Itoa(port), router); err != nil {
		log.Fatalln("ListenAndServe err:", err)
	}
}
