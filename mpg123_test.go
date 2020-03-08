package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetPlaybackInfoFromFrameLine(t *testing.T) {
	assert := assert.New(t)

	result, err := getPlaybackInfoFromFrameLine("@F 100 10 12.34 5.678")

	assert.EqualValues(100, result.Frame)
	assert.EqualValues(10, result.FramesRemaining)
	assert.EqualValues(12.34, result.Seconds)
	assert.EqualValues(5.678, result.SecondsRemaining)
	assert.Nil(err)
}

func TestGetPlaybackInfoFromFrameLineWithEmptyString(t *testing.T) {
	assert := assert.New(t)

	result, err := getPlaybackInfoFromFrameLine("")

	assert.EqualValues(0, result.Frame)
	assert.EqualValues(0, result.FramesRemaining)
	assert.EqualValues(0, result.Seconds)
	assert.EqualValues(0, result.SecondsRemaining)
	assert.Nil(err)
}
