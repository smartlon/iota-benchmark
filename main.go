package main

import (
	"github.com/smartlon/iota-benchmark/docker"
)

func main () {
	docker.Start(2)
	docker.Write()
}