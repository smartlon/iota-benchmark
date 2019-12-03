package main

import (
	"github.com/smartlon/iota-benchmark/docker"
	"time"
)

func main () {
	go docker.Start(2)
	time.Sleep(time.Duration(5)*time.Second)
	docker.Write()
}