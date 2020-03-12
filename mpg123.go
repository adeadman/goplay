package main

import (
	"bufio"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type PlaybackInfo struct {
	// Current frame
	Frame uint64
	// Number of remaining frames
	FramesRemaining uint64
	// Current position in seconds
	Seconds float64
	// Remaining seconds
	SecondsRemaining float64
}

type Mpg123Process struct {
	// the command of the subprocess being executed
	command *exec.Cmd
	// a channel where the process standard output will be buffered
	stdOut chan string
	// a channel which will feed into the process standard input
	stdIn chan string
}

func readFromPipe(pipe io.ReadCloser, channel chan string) {
	bufioReader := bufio.NewReader(pipe)
	for {
		output, _, err := bufioReader.ReadLine()
		if err != nil || err == io.EOF {
			break
		}
		log.Debug(fmt.Sprintf(">>>>>> STDOUT read: %s", output))
		channel <- string(output)
	}
}

func writeToPipe(pipe io.WriteCloser, channel chan string) {
	for {
		select {
		case line := <-channel:
			io.WriteString(pipe, line)
			log.Debug(fmt.Sprintf(">>>>>> STDIN written: %s", line))
			continue
		default:
		}
		time.Sleep(1 * time.Millisecond)
	}
}

func (process *Mpg123Process) init(cfg *MainConfig) (err error) {
	// start mpg123 in remote command mode with the -R flag
	process.command = exec.Command(cfg.PlayerPath, "-R")
	process.command.Dir = cfg.MusicDir

	// get a pipe to the process standard input and output
	stdInPipe, err := process.command.StdinPipe()
	if err != nil {
		return
	}
	stdOutPipe, err := process.command.StdoutPipe()
	if err != nil {
		return
	}

	process.stdOut = make(chan string)
	go readFromPipe(stdOutPipe, process.stdOut)

	process.stdIn = make(chan string)
	go writeToPipe(stdInPipe, process.stdIn)
	return
}

func (process *Mpg123Process) sendCommand(command string) {
	process.stdIn <- fmt.Sprintf("%s\n", command)
	// allow time for the application to read the input and respond
	time.Sleep(10 * time.Millisecond)
}

func (process *Mpg123Process) getOutputMatchingPrefix(prefix string) (lines []string) {
	for {
		select {
		case line := <-process.stdOut:
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

func (process *Mpg123Process) drainOutput() {
	_ = process.getOutputMatchingPrefix("")
}

func (process *Mpg123Process) start() (err error) {
	return process.command.Start()
}

// a frame line from mpg123 has the format:
//     @F <frame> <framesRemaining> <seconds> <secondsRemaining>
//
// where frame numbers are integer and seconds are decimals.
// For example:
//     @F 6467 250 168.93 6.53
//
// In this case, the player is at frame 6467 with 250 frames remaining, or at
// 168.93 seconds with 6.53 seconds of playtime remaining.
func getPlaybackInfoFromFrameLine(line string) (info *PlaybackInfo, err error) {
	info = &PlaybackInfo{}
	if line == "" {
		return
	}
	frameInfoLine := strings.Split(line, " ")
	// Start at position 1, as the first part is the `@F` prefix
	info.Frame, err = strconv.ParseUint(frameInfoLine[1], 10, 16)
	if err != nil {
		return
	}
	info.FramesRemaining, err = strconv.ParseUint(frameInfoLine[2], 10, 16)
	if err != nil {
		return
	}
	info.Seconds, err = strconv.ParseFloat(frameInfoLine[3], 64)
	if err != nil {
		return
	}
	info.SecondsRemaining, err = strconv.ParseFloat(frameInfoLine[4], 64)
	if err != nil {
		return
	}
	return
}
