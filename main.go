package main

import "flag"

var (
	configFilePath string
)

func init() {
	flag.StringVar(&configFilePath, "config", "config.toml", "filepath to config file")
}

func main() {
	flag.Parse()

	config := loadConfig(configFilePath)

	server := NewServer(config)
	server.loadConfig(config)
	server.Start()

	defer server.Close()
	server.Wait()
}
