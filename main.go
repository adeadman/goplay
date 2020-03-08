package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"html/template"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

var mpg123 *exec.Cmd
var mpg123Stdin io.WriteCloser
var mpg123Stdout io.ReadCloser
var mpg123Stderr io.ReadCloser
var mpg123StdoutChan chan string

func player(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path[1:] != "" {
		http.Error(w, "404 not found", http.StatusNotFound)
		return
	}
	t, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	t.Execute(w, nil)
}

func play(w http.ResponseWriter, r *http.Request) {
	io.WriteString(mpg123Stdin, "LOAD The Shadows - Apache 1963.mp3\n")
	fmt.Fprintf(w, "OK")
}

func pause(w http.ResponseWriter, r *http.Request) {
	io.WriteString(mpg123Stdin, "PAUSE\n")
	fmt.Fprintf(w, "OK")
}

func getPlaybackInfo(w http.ResponseWriter, r *http.Request) {
	// get the most recent frame info
	var recentStatusLine string
	for {
		select {
		case line := <-mpg123StdoutChan:
			if strings.HasPrefix(line, "@F ") {
				recentStatusLine = line
			}
			time.Sleep(1 * time.Millisecond)
			continue
		default:
		}
		break
	}
	playbackInfo, err := getPlaybackInfoFromFrameLine(recentStatusLine)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	playbackInfoJson, err := json.Marshal(playbackInfo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, string(playbackInfoJson))
}

func status(w http.ResponseWriter, r *http.Request) {
	// drain the stdout channel to remove any unread stdout lines
	for {
		select {
		case _ = <-mpg123StdoutChan:
			// we sleep for a ms to allow concurrent goroutines to write
			time.Sleep(1 * time.Millisecond)
			continue
		default:
		}
		break
	}
	io.WriteString(mpg123Stdin, "STATE\n")

	// allow time for the application to read the input and respond
	time.Sleep(30 * time.Millisecond)

	// collect all available output as above
	for {
		select {
		case line := <-mpg123StdoutChan:
			if strings.HasPrefix(line, "@STATE") {
				fmt.Fprintf(w, "%s\n", line)
			}
			time.Sleep(1 * time.Millisecond)
			continue
		default:
		}
		break
	}
	fmt.Fprintf(w, "OK")
}

func readFromPipe(pipe io.ReadCloser, channel chan string) {
	bufioReader := bufio.NewReader(pipe)
	for {
		output, _, err := bufioReader.ReadLine()
		if err != nil || err == io.EOF {
			break
		}
		channel <- string(output)
	}
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

	mpg123StdoutChan = make(chan string)
	go readFromPipe(mpg123Stdout, mpg123StdoutChan)

	err = mpg123.Start()
	if err != nil {
		log.Fatalf("mpg123 failed to start with '%s'\n", err)
	}

	log.Println(fmt.Sprintf(">>> Listening on %s", cfg.GetServerAddr()))
	http.HandleFunc("/", player)
	http.HandleFunc("/play", play)
	http.HandleFunc("/pause", pause)
	http.HandleFunc("/status", status)
	http.HandleFunc("/playbackInfo", getPlaybackInfo)
	log.Fatal(http.ListenAndServe(cfg.GetServerAddr(), nil))
}
