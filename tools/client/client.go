package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	port := flag.String("p", "4000", "Client port")
	flag.Parse()

	conn, err := net.Dial("tcp", ":"+*port)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Conneting to %s\n", conn.RemoteAddr())
	defer conn.Close()
	for {
		text, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			fmt.Println(err)
			continue
		}
		if strings.HasPrefix(text, "exit") {
			return
		}
		fmt.Fprintf(conn, text)
		message, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Println(err)
			break
		}
		fmt.Print("Message from server: " + message)
	}
}
