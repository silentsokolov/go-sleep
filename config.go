package main

import (
	"fmt"
	"log"
	"time"

	"github.com/BurntSushi/toml"
)

const (
	defaultAddress     = ":80"
	defaultBackendPost = 80
	defaultSleepAfter  = 20 * time.Minute
)

// Config ...
type Config struct {
	Port      string                `toml:"port"`
	SecretKey string                `toml:"secret_key"`
	Dummy     []*DummyConfig        `toml:"dummy"`
	GCE       []*GCEConfig          `toml:"gce"`
	EC2       []*EC2Config          `toml:"ec2"`
	AuthBasic map[string]*AuthGroup `toml:"auth"`
}

// AuthGroup ...
type AuthGroup struct {
	Users []string `toml:"users"`
}

// RouteConfig ...
type RouteConfig struct {
	Address      string               `toml:"address"`
	BackendPort  int                  `toml:"backend_port"`
	Hostnames    []string             `toml:"hostnames"`
	AuthGroup    string               `toml:"auth_group"`
	Certificates []*CertificateConfig `toml:"certificate"`
}

// String ...
func (c *RouteConfig) String() string {
	return fmt.Sprintf("%s:%s", c.Address, c.Hostnames[0])
}

// CertificateConfig ...
type CertificateConfig struct {
	CertFile string `toml:"cert_file"`
	KeyFile  string `toml:"key_file"`
}

// BaseConfig ...
type BaseConfig struct {
	SleepAfter    int64          `toml:"sleep_after"`
	UseInternalIP bool           `toml:"use_internal_ip"`
	Routes        []*RouteConfig `toml:"route"`
}

// GCEConfig ...
type GCEConfig struct {
	BaseConfig
	JWTPath   string `toml:"jwt_path"`
	ProjectID string `toml:"project_id"`
	Zone      string `toml:"zone"`
	Name      string `toml:"name"`
}

// EC2Config ...
type EC2Config struct {
	BaseConfig
	AccessKeyID     string `toml:"access_key_id"`
	SecretAccessKey string `toml:"secret_access_key"`
	Region          string `toml:"region"`
	InstanceID      string `toml:"instance_id"`
}

// DummyConfig ...
type DummyConfig struct {
	BaseConfig
	APIKey        string `toml:"api_key"`
	DummyID       string `toml:"dummy_id"`
	UseInternalIP bool   `toml:"use_internal_ip"`
}

func loadConfig(filepath string) *Config {
	var config Config

	if _, err := toml.DecodeFile(filepath, &config); err != nil {
		log.Fatal(err)
	}

	return &config
}
