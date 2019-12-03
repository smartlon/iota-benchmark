package main

import (
	"github.com/smartlon/iota-benchmark/docker"
	"time"
)

func main () {
	stop := make(chan bool)
	go docker.Start(stop)
	time.Sleep(time.Duration(2)*time.Second)
	stop <- true
	<-stop
	docker.Write()
}