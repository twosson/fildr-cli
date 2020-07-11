package main

import (
	"fil-pusher/internal/command"
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
	//instance := collector.GetInstance("node")
	//instance.SetJob("test")
	//instance.SetInstance("aaabc")
	//for range time.Tick(time.Second * 10) {
	//	fmt.Println("print metrics ...")
	//	fmt.Println(instance.GetMetrics())
	//}
}
