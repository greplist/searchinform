package main

import (
	"log"
	"net/http"
)

func CountryByIP(w http.ResponseWriter, r *http.Request) {

}

func main() {
	router := http.NewServeMux()
	router.HandleFunc("/api/country", CountryByIP)

	log.Println("Server start...")
	if err := http.ListenAndServe(":80", router); err != nil {
		log.Fatalln("ListenAndServe err:", err)
	}
}
