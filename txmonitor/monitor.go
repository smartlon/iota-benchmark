package txmonitor

import (
    "sync"
    "time"
    "fmt"
    "github.com/smartlon/iota-benchmark/txmonitor/transactions"
    "github.com/pebbe/zmq4"

)

type Config struct {
    Zmq_Address string
    Interval    int
}

var totalTime int64 = 0

func Monitor(wg1 *sync.WaitGroup) {
    // print out current zmq version
    major, minor, patch := zmq4.Version()
    fmt.Printf("running ZMQ %d.%d.%d\n", major, minor, patch)

    // read config file
    //fileBytes, err := ioutil.ReadFile("./config.json")
    //if err != nil {
    //    panic(err)
    //}
    //if err := json.Unmarshal(fileBytes, &value); err != nil {
    //    panic(err)
    //}
    value := Config{}
    value.Zmq_Address = "tcp://202.117.43.212:5556"
    value.Interval = 5
    fmt.Printf("ZMQ addr: %s\n", value.Zmq_Address)
    fmt.Printf("Interval: %d\n", value.Interval)
    var wg sync.WaitGroup
    // start feeds
    go transactions.StartTxFeed(value.Zmq_Address)
    go transactions.StartMilestoneFeed(value.Zmq_Address)
    go transactions.StartConfirmationFeed(value.Zmq_Address)
    go transactions.StartDoubleFeed(value.Zmq_Address)
    wg.Add(1)
    time.Sleep(time.Second*time.Duration(value.Interval))
    go transactions.StartLog(value.Interval,&wg)
    wg.Wait()
    wg1.Done()
}