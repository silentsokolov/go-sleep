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
