package spammer


import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/jroimartin/gocui"
	"github.com/pebbe/zmq4"
	"io/ioutil"
	"strings"
	"sync/atomic"
	"text/tabwriter"
	"time"
)

type config struct {
	URIs []string `json:"uris"`
}

var next = make(chan struct{})
var nodes []node
var cnf *config

func main() {
	flag.Parse()

	// read in config
	configBytes, err := ioutil.ReadFile("./config.json")
	must(err)

	cnf = &config{}
	must(json.Unmarshal(configBytes, cnf))

	g, err := gocui.NewGui(gocui.OutputNormal)
	must(err)
	defer g.Close()

	g.SetManagerFunc(layout)

	must(g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return gocui.ErrQuit
	}))

	// init nodes
	nodes = make([]node, len(cnf.URIs))
	for i, uri := range cnf.URIs {
		nodes[i] = node{uri: uri, received: make(map[string]struct{})}
		go nodes[i].stream()
	}

	go func() {
		for {
			<-time.After(time.Duration(500) * time.Millisecond)
			g.Update(layout)
		}
	}()

	g.MainLoop()
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	view, err := g.SetView("main", 0, 0, maxX, maxY)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		writeNodesInfo(view)
	}
	view.Clear()
	writeNodesInfo(view)

	return nil
}

func writeNodesInfo(view *gocui.View) {
	var fastestRec, slowestRec int64
	var fastest, slowest string
	var highest, lowest int
	w := tabwriter.NewWriter(view, 0, 0, 4, ' ', tabwriter.Debug)
	for _, n := range nodes {
		if n.lastRec == 0 {
			continue
		}
		if n.lastRec <= fastestRec || fastestRec == 0 {
			fastestRec = n.lastRec
			fastest = n.uri
		} else {
			slowestRec = n.lastRec
			slowest = n.uri
		}
		received := len(n.received)
		if received > highest || highest == 0 {
			highest = received
		}
		if received < lowest || lowest == 0 {
			lowest = received
		}
		fmt.Fprintf(w, "%s\t received %d\t last tx %s\t %d tx/s\n", n.uri, len(n.received), n.last[:5], n.tps)
	}
	fmt.Fprintf(w, "fastest %s, slowest %s (%dms latency)\n", fastest, slowest, slowestRec-fastestRec)
}

type node struct {
	uri      string
	received map[string]struct{}
	last     string
	lastRec  int64
	tps      int64
}

func (n *node) stream() {
	socket, err := zmq4.NewSocket(zmq4.SUB)
	must(err)
	must(socket.SetSubscribe("tx"))
	err = socket.Connect(n.uri)
	must(err)

	var passed int64

	go func() {
		for {
			<-time.After(time.Duration(1) * time.Second)
			n.tps = atomic.LoadInt64(&passed)
			passed = 0
		}
	}()

	for {
		msg, err := socket.Recv(0)
		must(err)
		split := strings.Split(msg, " ")
		if len(split) != 13 {
			continue
		}

		n.last = split[1]
		n.lastRec = time.Now().UnixNano() / int64(time.Millisecond)
		n.received[split[1]] = struct{}{}
		atomic.AddInt64(&passed, 1)
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}