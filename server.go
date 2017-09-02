package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"
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

// CountryByIP ..
func (ctrl *Controller) CountryByIP(w http.ResponseWriter, r *http.Request) {

}

func main() {
	flag.Parse()

	log.Println("Parsing config file", configPath)
	conf, err := ParseConfig(configPath)
	if err != nil {
		log.Fatalln("Parse err:", err)
	}

	ctrl := NewController(conf)

	router := http.NewServeMux()
	router.HandleFunc("/api/country", ctrl.CountryByIP)

	port := conf.HTTP.Port
	log.Printf("Server start on %v port...\n", port)
	if err := http.ListenAndServe(":"+strconv.Itoa(port), router); err != nil {
		log.Fatalln("ListenAndServe err:", err)
	}
}
