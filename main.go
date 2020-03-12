package main

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"html/template"
	"net/http"
)

var mpg123 *Mpg123Process

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
	mpg123.sendCommand("LOAD The Shadows - Apache 1963.mp3")
	fmt.Fprintf(w, "OK")
}

func stop(w http.ResponseWriter, r *http.Request) {
	mpg123.sendCommand("STOP")
	mpg123.drainOutput()
	fmt.Fprintf(w, "OK")
}

func pause(w http.ResponseWriter, r *http.Request) {
	mpg123.drainOutput()
	mpg123.sendCommand("PAUSE")
	pauseLines := mpg123.getOutputMatchingPrefix("@P ")
	if len(pauseLines) > 0 {
		if pauseLines[0] == "@P 0" {
			fmt.Fprintf(w, "NOT PLAYING")
		} else if pauseLines[0] == "@P 2" {
			fmt.Fprintf(w, "RESUMED")
		} else {
			fmt.Fprintf(w, "?????")
		}
	} else {
		fmt.Fprintf(w, "PAUSED")
	}
}

func getPlaybackInfo(w http.ResponseWriter, r *http.Request) {
	// get the most recent frame info
	statusLines := mpg123.getOutputMatchingPrefix("@F ")
	recentStatusLine := ""
	if len(statusLines) > 0 {
		recentStatusLine = statusLines[len(statusLines)-1]
	}

	// create JSON representation of frame info line
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
	mpg123.drainOutput()
	mpg123.sendCommand("STATE")

	stateLines := mpg123.getOutputMatchingPrefix("@STATE ")
	for _, line := range stateLines {
		fmt.Fprintf(w, fmt.Sprintf("%s\n", line))
	}
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
	mpg123 = &Mpg123Process{}
	err = mpg123.init(cfg)
	if err != nil {
		log.Fatal(fmt.Sprintf("!!! Failed to initialise mpg123: %v", err))
	}
	err = mpg123.start()
	if err != nil {
		log.Fatal(fmt.Sprintf("!!! Failed to start mpg123: %v", err))
	}

	log.Println(fmt.Sprintf(">>> Listening on %s", cfg.GetServerAddr()))
	http.HandleFunc("/", player)
	http.HandleFunc("/play", play)
	http.HandleFunc("/stop", stop)
	http.HandleFunc("/pause", pause)
	http.HandleFunc("/status", status)
	http.HandleFunc("/playbackInfo", getPlaybackInfo)
	log.Fatal(http.ListenAndServe(cfg.GetServerAddr(), nil))
}
