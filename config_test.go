package main

import (
	"io/ioutil"
	"os"
	"testing"
)

var exampleConfigFile = `
port = ":9090"
secret_key = "secret"

[auth]
[auth.admins]
users = ["test:$apr1$bfLZ0ZMK$CYhTBqS.Yl.V1hbOpHze51"]
`

func initTestConfigFile() *os.File {
	content := []byte(exampleConfigFile)
	tmpfile, err := ioutil.TempFile("", "example")
	if err != nil {
		panic(err)
	}

	if _, err := tmpfile.Write(content); err != nil {
		panic(err)
	}

	if err := tmpfile.Close(); err != nil {
		panic(err)
	}

	return tmpfile
}

func TestRouteConfig_String(t *testing.T) {
	s := "[example.com example.com] on :80"
	inst := &RouteConfig{
		Address:   ":80",
		Hostnames: []string{"example.com", "example.com"},
	}

	if inst.String() != s {
		t.Errorf("RouteConfig.String returned %+v, want %+v", inst.String(), s)
	}
}

func TestLoadConfig(t *testing.T) {
	tmpfile := initTestConfigFile()
	defer os.Remove(tmpfile.Name())

	config := loadConfig(tmpfile.Name())

	if config.Port != ":9090" {
		t.Errorf("loadConfig params returned %+v, want %+v", config.Port, ":9090")
	}

	if config.SecretKey != "secret" {
		t.Errorf("loadConfig params returned %+v, want %+v", config.Port, "secret")
	}

	if _, ok := config.AuthBasic["admins"]; !ok {
		t.Error("loadConfig auth not load auth config")
	}
}
