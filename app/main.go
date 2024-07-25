package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	var dirname string
	for i, arg := range os.Args {
		if arg == "--directory" && i+1 < len(os.Args) {
			dirname = os.Args[i+1]
		}
	}

	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Failed to accept connection")
			os.Exit(1)
		}

		go handlecConnection(conn, dirname)
	}
}
