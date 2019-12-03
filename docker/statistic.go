package docker

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

const INTERVAL = 1

var logStats map[string][]Stats
func Start(duration int) {
	var wg sync.WaitGroup
	logStats = make(map[string][]Stats,0)
	timeout := time.After(time.Second * time.Duration(duration))
	wg.Add(1)
	go func() {
		for {
			select {
				case <-timeout:
					fmt.Println("timeout")
					return
				default:
					time.AfterFunc(time.Duration(INTERVAL) * time.Millisecond,func() {
						stats,err := DefaultCommunicator.Stats()
						if err != nil {
							return
						}
						for _,stat := range stats {
							logStats[stat.Container] = append(logStats[stat.Container],stat)
						}
					})
			}
		}
		wg.Done()
	}()
	wg.Wait()
}

func Write(){
	filename := "output.txt"
	if len(os.Args) > 1 {
		filename = os.Args[1]
	}
	f, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	log.Printf("Writing output to '%v'", f.Name())
	for _,logStat := range logStats {
		var out string
		for _, stat := range logStat {
			out = fmt.Sprintf("%v%v: %v\n", out, time.Now(), stat)
		}
		if _, err := f.WriteString(out); err != nil {
			panic(err)
		}
	}
}
