package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"

	"searchinform/cache"
)

var (
	configPath string
)

func init() {
	flag.StringVar(&configPath, "c", "conf.json", "config filepath")
}

// Controller - main struct with all dependences
type Controller struct {
	client Client
	logger log.Logger
}

// NewController - constructor for Controller struct
func NewController(conf *Config) *Controller {
	return &Controller{
		client: *NewClient(conf),
		logger: *log.New(os.Stdout, conf.Log.Prefix, conf.LogFlags()),
	}
}

// Init run all background jobs
func (ctrl *Controller) Init() {
	go cache.Cleaner(context.Background(), &ctrl.client.cache)
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

// CountryByIP ..
func (ctrl *Controller) CountryByIP(w http.ResponseWriter, r *http.Request) {
	host := r.FormValue("host")
	if host == "" {
		host = r.Host
	}

	ip, err := lookup(host)
	if err != nil {
		ctrl.error(w, "Lookup err: "+err.Error(), http.StatusBadRequest)
		return
	}

	country, err := ctrl.client.Resolve(ip)
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

	ctrl := NewController(conf)
	ctrl.Init()

	router := http.NewServeMux()
	router.HandleFunc("/api/country", ctrl.CountryByIP)

	port := conf.HTTP.Port
	log.Printf("Server start on %v port...\n", port)
	if err := http.ListenAndServe(":"+strconv.Itoa(port), router); err != nil {
		log.Fatalln("ListenAndServe err:", err)
	}
}
