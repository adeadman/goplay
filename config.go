package main

import (
	"fmt"
	"github.com/spf13/viper"
)

// Defaults
const (
	goplayEnvPrefix       string = "GOPLAY"
	defaultServerHostname        = "localhost"
	defaultServerPort     uint16 = 8080
)

// Env variable mapping
const (
	envTypeEnv        string = "ENV"
	ServerHostnameEnv        = "HOST"
	ServerPortEnv            = "PORT"
)

// Struct holding basic configuration for application
type MainConfig struct {
	// Type of environment
	EnvType string
	// local hostname or IP address
	ServerHostname string
	// Application port
	ServerPort uint16
}

func (cfg *MainConfig) GetServerAddr() string {
	return fmt.Sprintf("%s:%d", cfg.ServerHostname, cfg.ServerPort)
}

func initConfig() (cfg *MainConfig) {
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
	return
}
