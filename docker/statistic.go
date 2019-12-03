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
func Start(duration int,wg1 *sync.WaitGroup) {
	logStats = make(map[string][]Stats,0)
	var wg *sync.WaitGroup
	wg = &sync.WaitGroup{}
	count :=0
	for range time.Tick(time.Duration(INTERVAL) * time.Millisecond) {
		if count>=duration {
			break
		}
		wg.Add(1)
		go func() {
			stats, err := DefaultCommunicator.Stats()
			if err != nil {
				return
			}
			for _, stat := range stats {
				logStats[stat.Container] = append(logStats[stat.Container], stat)
			}
			wg.Done()
		}()
		count++
	}
	wg1.Done()
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
