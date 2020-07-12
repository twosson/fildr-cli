package main

import (
	"fildr-cli/internal/command"
	"math/rand"
	"time"
)

var (
	version   = "(dev-version)"
	gitCommit = "(dev-gitCommit)"
	buildTime = "(dev-buildTime)"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	command.Execute(version, gitCommit, buildTime)
}
