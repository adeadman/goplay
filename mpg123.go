package main

import (
	"strconv"
	"strings"
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
	info.Seconds, err = strconv.ParseFloat(frameInfoLine[3], 32)
	if err != nil {
		return
	}
	info.SecondsRemaining, err = strconv.ParseFloat(frameInfoLine[4], 32)
	if err != nil {
		return
	}
	return
}
