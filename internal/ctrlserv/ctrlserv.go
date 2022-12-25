package ctrlserv

import (
	"bufio"
	"log"
	"net"
	"regexp"
)

func RunControlServer(port string, mainControlCmd chan string, repeaterControlChannels *[]chan string) {
	shutdownCommand := "shutdown"
	validCommands := []string{shutdownCommand, "refresh"}
	endLineMatcher := regexp.MustCompile(`\n|\r|\r\n`)

	listener, err := net.Listen("tcp", "localhost:"+port)
	if err != nil {
		panic(err)
	}
	log.Printf("Control server running on port: %s", port)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		cmd, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			log.Println(err)
			continue
		}
		cmd = endLineMatcher.ReplaceAllString(cmd, "")
		log.Printf("Received command: %s", cmd)

		validCmd := false
		for _, c := range validCommands {
			if c == cmd {
				validCmd = true
				break
			}
		}
		if validCmd {
			for _, c := range *repeaterControlChannels {
				log.Printf("Send %s command to %v", cmd, c)
				c <- cmd
			}
			if cmd == shutdownCommand {
				log.Printf("Send %s command to main routine", cmd)
				mainControlCmd <- cmd
			}
			conn.Write([]byte("OK\n"))
		} else {
			log.Printf("Incorrect command: %s", cmd)
			conn.Write([]byte("ERR\n"))
		}
		conn.Close()
	}
}
