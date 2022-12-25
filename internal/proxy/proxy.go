package proxy

import (
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

type Repeater struct {
	Address    string
	Port       uint16
	SourcePort uint16
}

type ProxyConfig struct {
	Listener           net.Listener
	Destination        string
	Repeaters          []Repeater
	DestinationTimeout time.Duration
	RepeaterTimeout    time.Duration
	ControlChannels    *[]chan string
}

func ProxyOriginalPort(config ProxyConfig) {
	for {
		sourceConn, err := config.Listener.Accept()
		if err != nil {
			panic(err)
		}
		log.Printf("Connection accepted from: %v", sourceConn.RemoteAddr())
		controlCmd := make(chan string)
		*config.ControlChannels = append(*config.ControlChannels, controlCmd)
		log.Printf("Create repeater control channel: %v", controlCmd)
		go handleConnection(sourceConn, config, controlCmd)
	}
}

func handleConnection(sourceConn net.Conn, config ProxyConfig, controlCmd chan string) {
	destConn, err := net.DialTimeout("tcp", config.Destination, config.DestinationTimeout)
	if err != nil {
		log.Panicf("Can't connect to destination host %s. Error: %v", config.Destination, err)
	}
	log.Printf("Connected to original destination: %s", config.Destination)

	repeaterConns := refreshRepeaterConns(config)

	go func() {
		errorCounter := 0
		buf := make([]byte, 32*1024)
		var err error
		for {
			nr, er := sourceConn.Read(buf)
			if nr > 0 {
				wr := writeData(destConn, buf, nr)
				if wr != nil {
					err = wr
					break
				}
				for _, r := range repeaterConns {
					wrr := writeData(r, buf, nr)
					if wrr != nil {
						log.Println(wrr)
						errorCounter++
					}
				}
				if errorCounter > 4 {
					errorCounter = 0
					controlCmd <- "refresh"
				}
			}
			if er != nil {
				if er != io.EOF {
					err = er
				}
				break
			}
		}
		if err != nil {
			log.Println(err)
			return
		}
	}()
	go func() {
		_, err := io.Copy(sourceConn, destConn)
		defer sourceConn.Close()
		if err != nil {
			log.Println(err)
			return
		}
	}()

	for {
		cmd := <-controlCmd
		switch cmd {
		case "refresh":
			for _, r := range repeaterConns {
				r.Close()
			}
			repeaterConns = refreshRepeaterConns(config)
		case "shutdown":
			log.Printf("Closing %s routine...", sourceConn.RemoteAddr())
			return
		}
	}
}

func refreshRepeaterConns(config ProxyConfig) []net.Conn {
	var repeaterConns []net.Conn
	for _, r := range config.Repeaters {
		repeaterAddress := fmt.Sprintf("%s:%d", r.Address, r.Port)
		repeaterConn, err := net.DialTimeout("tcp", repeaterAddress, config.RepeaterTimeout)
		if err != nil {
			log.Printf("Can't connect to repeater host %s. Error: %v", repeaterAddress, err)
			continue
		}
		log.Printf("Connected to repeater: %s", repeaterAddress)
		repeaterConns = append(repeaterConns, repeaterConn)
	}
	if len(repeaterConns) == 0 {
		log.Println("There is no active repeaters. Just forward original traffic")
	} else {
		log.Printf("Forward original traffic and repeat it to %d hosts", len(repeaterConns))
	}
	return repeaterConns
}

func writeData(writer io.Writer, buf []byte, n int) (err error) {
	nw, ew := writer.Write(buf[0:n])
	if ew != nil {
		err = ew
	}
	if n != nw {
		err = fmt.Errorf("write error, expected to write %d bytes, actually %d", n, nw)
	}
	return
}
