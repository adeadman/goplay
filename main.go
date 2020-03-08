package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os/exec"
)

//var mpg123StdoutBuf, mpg123StderrBuf bytes.Buffer
var mpg123 *exec.Cmd
var mpg123Stdin io.WriteCloser
var mpg123Stdout io.ReadCloser
var mpg123Stderr io.ReadCloser

func player(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path[1:] != "" {
		http.Error(w, "404 not found", http.StatusNotFound)
		return
	}
	fmt.Fprintf(w, "Player will be loaded here")
}

func play(w http.ResponseWriter, r *http.Request) {
	io.WriteString(mpg123Stdin, "LOAD The Shadows - Apache 1963.mp3\n")
	fmt.Fprintf(w, "OK")
}

func pause(w http.ResponseWriter, r *http.Request) {
	io.WriteString(mpg123Stdin, "PAUSE\n")
	fmt.Fprintf(w, "OK")
}

func main() {
	log.Println(">>> Loading config")
	cfg, err := initConfig()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(fmt.Sprintf(">>> mpg123 found at %s", cfg.PlayerPath))
	log.Println(fmt.Sprintf(">>> working directory: %s", cfg.MusicDir))
	log.Println(">>> Starting mpg123 in Remote Command mode with attached pipes")

	mpg123 = exec.Command(cfg.PlayerPath, "-R")
	mpg123.Dir = cfg.MusicDir
	mpg123Stdin, err = mpg123.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}
	mpg123Stdout, err = mpg123.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	mpg123Stderr, err = mpg123.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	err = mpg123.Start()
	if err != nil {
		log.Fatalf("mpg123 failed to start with '%s'\n", err)
	}

	log.Println(fmt.Sprintf(">>> Listening on %s", cfg.GetServerAddr()))
	http.HandleFunc("/", player)
	http.HandleFunc("/play", play)
	http.HandleFunc("/pause", pause)
	log.Fatal(http.ListenAndServe(cfg.GetServerAddr(), nil))
}
