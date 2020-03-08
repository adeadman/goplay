package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
)

var afplay_pid int

func player(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, %s", r.URL.Path[1:])
}

func main() {
	log.Println(">>> Loading config")
	cfg := initConfig()
	log.Println(fmt.Sprintf(">>> Listening on %s", cfg.GetServerAddr()))
	http.HandleFunc("/", player)
	log.Fatal(http.ListenAndServe(cfg.GetServerAddr(), nil))
}
