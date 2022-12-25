package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
)

func main() {
	port := flag.String("p", "4000", "Server port")
	flag.Parse()

	ln, err := net.Listen("tcp", ":"+*port)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Listening %s port...\n", *port)
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Printf("Client accepted from %v\n", conn.RemoteAddr())
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	for {
		message, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Printf("Client %v disconnected\n", conn.RemoteAddr())
			} else {
				fmt.Println(err)
			}
			return
		}
		fmt.Print(message)
		conn.Write([]byte(message))
	}
}
