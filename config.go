package main

import (
	"errors"
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"os/exec"
)

// Defaults
const (
	goplayEnvPrefix       string = "GOPLAY"
	defaultMusicDir              = "~/Music/"
	defaultServerHostname        = "localhost"
	defaultServerPort     uint16 = 8080
)

// Env variable mapping
const (
	envTypeEnv        string = "ENV"
	ServerHostnameEnv        = "HOST"
	ServerPortEnv            = "PORT"
	MusicDirEnv              = "DIR"
)

// Struct holding basic configuration for application
type MainConfig struct {
	// Type of environment
	EnvType string
	// local hostname or IP address
	ServerHostname string
	// Application port
	ServerPort uint16
	// Working Directory for Music (root)
	MusicDir string
	// Path to mpg123 binary
	PlayerPath string
}

func (cfg *MainConfig) GetServerAddr() string {
	return fmt.Sprintf("%s:%d", cfg.ServerHostname, cfg.ServerPort)
}

func (cfg *MainConfig) getPlayer() (err error) {
	// check if mpg123 is on the path
	cfg.PlayerPath, err = exec.LookPath("mpg123")
	if err != nil {
		return errors.New("Unable to find mpg123 - please ensure it is installed and in your path")
	}
	return
}

func initConfig() (cfg *MainConfig, err error) {
	viper.SetEnvPrefix(goplayEnvPrefix)

	cfg = &MainConfig{}
	cfg.EnvType = viper.GetString(envTypeEnv)

	viper.AutomaticEnv()

	cfg.ServerHostname = viper.GetString(ServerHostnameEnv)
	if cfg.ServerHostname == "" {
		cfg.ServerHostname = defaultServerHostname
	}
	cfg.ServerPort = uint16(viper.GetInt(ServerPortEnv))
	if cfg.ServerPort == 0 {
		cfg.ServerPort = defaultServerPort
	}
	cfg.MusicDir = viper.GetString(MusicDirEnv)
	if cfg.MusicDir == "" {
		cfg.MusicDir = defaultMusicDir
	}
	// expand any homedir in the path
	cfg.MusicDir, err = homedir.Expand(cfg.MusicDir)
	if err != nil {
		return
	}
	// Check Player is available
	err = cfg.getPlayer()
	if err != nil {
		return
	}
	return
}
