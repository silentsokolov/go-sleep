package main

import (
	"flag"
	"strings"

	"github.com/Sirupsen/logrus"

	"github.com/silentsokolov/go-sleep/log"
)

var (
	configFilePath string
)

func init() {
	flag.StringVar(&configFilePath, "config", "config.toml", "filepath to config file")
}

func main() {
	flag.Parse()
	config := loadConfig(configFilePath)

	level, err := logrus.ParseLevel(strings.ToLower(config.LogLevel))
	if err != nil {
		log.Error("Error getting level", err)
	}
	log.SetLevel(level)

	server := NewServer(config)
	server.loadConfig(config)
	server.Start()

	defer server.Close()
	server.Wait()
}
