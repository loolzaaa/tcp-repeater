package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/loolzaaa/tcp-repeater/internal/ctrlserv"
	"github.com/loolzaaa/tcp-repeater/internal/proxy"
)

type sourcePorts []uint16
type repeaterFlags []string

func (flags *sourcePorts) String() string {
	return fmt.Sprint(*flags)
}

func (flags *sourcePorts) Set(value string) error {
	i, err := strconv.Atoi(value)
	if err != nil {
		panic(err)
	}
	*flags = append(*flags, uint16(i))
	return nil
}

func (flags *repeaterFlags) String() string {
	return fmt.Sprint(*flags)
}

func (flags *repeaterFlags) Set(value string) error {
	*flags = append(*flags, value)
	return nil
}

func parseRepeaters(flags repeaterFlags) []proxy.Repeater {
	var repeaters []proxy.Repeater
	for _, flag := range flags {
		parts := strings.Split(flag, "-")
		repeaterParts := strings.Split(parts[0], ":")
		address := repeaterParts[0]
		port, err := strconv.Atoi(repeaterParts[1])
		if err != nil {
			panic(err)
		}
		sourcePort, err := strconv.Atoi(parts[1])
		if err != nil {
			panic(err)
		}
		repeaters = append(repeaters, proxy.Repeater{Address: address, Port: uint16(port), SourcePort: uint16(sourcePort)})
	}
	return repeaters
}

func main() {
	log.Println("Starting repeater service...")

	var destination string
	var sourcePorts sourcePorts
	var repeaterFlags repeaterFlags
	var destTimeout time.Duration
	var repeaterTimeout time.Duration
	var controlServerPort string
	var testPort int

	flag.StringVar(&destination, "d", "", "Destination host")
	flag.Var(&sourcePorts, "p", "Source ports")
	flag.Var(&repeaterFlags, "r", "Repeaters pattern")
	flag.DurationVar(&destTimeout, "td", 250*time.Millisecond, "Destination connect timeout")
	flag.DurationVar(&repeaterTimeout, "tr", 250*time.Millisecond, "Repeater connect timeout")
	flag.StringVar(&controlServerPort, "cp", "6400", "Control server port")
	flag.IntVar(&testPort, "test-port", -1, "Test port for forwarded destination")
	flag.Parse()

	if destination == "" {
		log.Panic("You must provide a destination host")
	}
	if len(sourcePorts) < 1 {
		log.Panic("You must provide at least 1 source port")
	}
	log.Printf("Destination connect timeout: %v", destTimeout)
	log.Printf("Repeater connect timeout: %v", repeaterTimeout)

	log.Println("Parse repeater patterns...")
	repeaters := parseRepeaters(repeaterFlags)
	if len(repeaters) > 0 {
		log.Println("Current repeaters:")
		for i, r := range repeaters {
			log.Printf("%d: %v", i, r)
		}
	} else {
		log.Println("There is no repeater patterns in arguments")
	}

	// Run main repeater routine
	log.Printf("Try to forward traffic for %s from ports %v", destination, sourcePorts)
	listeners := make(map[uint16]net.Listener)
	for _, p := range sourcePorts {
		listener, err := net.Listen("tcp", "localhost:"+fmt.Sprint(p))
		if err != nil {
			panic(err)
		}
		log.Printf("Listen port: %d", p)
		listeners[p] = listener
	}

	repeaterCtrlChannels := make([]chan string, 0)
	for port, l := range listeners {
		var filteredRepeaters []proxy.Repeater
		for _, r := range repeaters {
			if r.SourcePort == port {
				filteredRepeaters = append(filteredRepeaters, r)
			}
		}
		var destAddr string
		if testPort <= 0 {
			destAddr = fmt.Sprintf("%s:%d", destination, port)
		} else {
			log.Printf("Test port active. All connections forward to %d port", testPort)
			destAddr = fmt.Sprintf("%s:%d", destination, testPort)
		}
		config := proxy.ProxyConfig{Listener: l, Destination: destAddr, Repeaters: filteredRepeaters, DestinationTimeout: destTimeout, RepeaterTimeout: repeaterTimeout, ControlChannels: &repeaterCtrlChannels}
		go proxy.ProxyOriginalPort(config)
	}

	// Run control server
	mainControlCmd := make(chan string)
	go ctrlserv.RunControlServer(controlServerPort, mainControlCmd, &repeaterCtrlChannels)

	// Await shutdown comman
	<-mainControlCmd
	log.Println("Repeater service stopped")
}
