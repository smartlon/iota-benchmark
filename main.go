package main

import (
	"github.com/smartlon/iota-benchmark/docker"
	"sync"
	"time"
)

func main () {
	var wg1 *sync.WaitGroup
	wg1 = sync.WaitGroup{}
	wg1.Add(1)
	go docker.Start(2,wg1)
	wg1.Wait()
	docker.Write()
}