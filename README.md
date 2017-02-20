[![Build Status](https://travis-ci.org/silentsokolov/go-sleep.svg?branch=master)](https://travis-ci.org/silentsokolov/go-sleep)
[![GoDoc](https://godoc.org/github.com/silentsokolov/go-sleep?status.svg)](https://godoc.org/github.com/silentsokolov/go-sleep/) [![codecov](https://codecov.io/gh/silentsokolov/go-sleep/branch/master/graph/badge.svg)](https://codecov.io/gh/silentsokolov/go-sleep)
[![Go Report Card](https://goreportcard.com/badge/github.com/silentsokolov/go-sleep)](https://goreportcard.com/report/github.com/silentsokolov/go-sleep)

# go-sleep

go-sleep helps to automatically start cloud instances [Google Compute Engine](https://cloud.google.com/compute/) / [Amazon EC2](https://aws.amazon.com/ec2/) by request (HTTP) and stopping unused instances, after some time.

Here is basic workflow: the go-sleep handles all incoming requests, if instance running, proxy all traffic. Else request start instance and waiting him.

![Diagram](https://raw.githubusercontent.com/silentsokolov/go-sleep/master/docs/diagram.png)


## Installation

Download latest binary from https://github.com/silentsokolov/go-sleep/releases


## Getting started

Run `./go-sleep -config=/path/to/config.toml`


## Config

### global

```toml
# Port
# Reserved for web API interface
port = ":9090"

# Secret key
# Is passed along with every request to that site in the X-Go-Sleep-Key header
secret_key = ""

# Log level
log_level = "info"
```

### Basic auth

```toml
# Group user for basic auth
# Passwords can be encoded in MD5, SHA1 and BCrypt: you can use htpasswd to generate those ones

# [auth]
#  [auth.<group_name>]
#    users = ["<user>:<password>", "<user>:<password>"]

# This example register two groups admins/freelancers with user "test" with password "test"
[auth]
  [auth.admins]
    users = ["test:$apr1$bfLZ0ZMK$CYhTBqS.Yl.V1hbOpHze51"]
  [auth.freelancers]
    users = ["test:$apr1$bfLZ0ZMK$CYhTBqS.Yl.V1hbOpHze51"]
```
