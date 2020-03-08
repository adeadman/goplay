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

func getLinesFromStdoutMatchingPrefix(prefix string) (lines []string) {
	for {
		select {
		case line := <-mpg123StdoutChan:
			if strings.HasPrefix(line, prefix) {
				lines = append(lines, line)
			}
			time.Sleep(1 * time.Millisecond)
			continue
		default:
		}
		break
	}
	return
}

func drainStdout() {
	// drain the stdout channel to remove any unread stdout lines
	_ = getLinesFromStdoutMatchingPrefix("")
}

func sendCommand(command string) {
	io.WriteString(mpg123Stdin, fmt.Sprintf("%s\n", command))
	// allow time for the application to read the input and respond
	time.Sleep(10 * time.Millisecond)
}

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
	sendCommand("LOAD The Shadows - Apache 1963.mp3")
	fmt.Fprintf(w, "OK")
}

func stop(w http.ResponseWriter, r *http.Request) {
	sendCommand("STOP")
	drainStdout()
	fmt.Fprintf(w, "OK")
}

func pause(w http.ResponseWriter, r *http.Request) {
	drainStdout()
	sendCommand("PAUSE")
	pauseLines := getLinesFromStdoutMatchingPrefix("@P ")
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
	statusLines := getLinesFromStdoutMatchingPrefix("@F ")
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
	drainStdout()
	sendCommand("STATE")

	stateLines := getLinesFromStdoutMatchingPrefix("@STATE ")
	for _, line := range stateLines {
		fmt.Fprintf(w, fmt.Sprintf("%s\n", line))
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
	http.HandleFunc("/stop", stop)
	http.HandleFunc("/pause", pause)
	http.HandleFunc("/status", status)
	http.HandleFunc("/playbackInfo", getPlaybackInfo)
	log.Fatal(http.ListenAndServe(cfg.GetServerAddr(), nil))
}
