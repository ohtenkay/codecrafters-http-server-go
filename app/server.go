package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
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

		go handlecConnection(conn)
	}
}

func handlecConnection(conn net.Conn) {
	defer conn.Close()

	b := make([]byte, 1024)
	conn.Read(b)

	lineHeaders, _ := splitByFirstOccurrence(string(b), "\r\n\r\n")
	line, headers := splitByFirstOccurrence(lineHeaders, "\r\n")
	lineParts := strings.Split(line, " ")
	urlParts := strings.Split(lineParts[1], "/")

	switch urlParts[1] {
	case "":
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))

	case "echo":
		conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: " + fmt.Sprintf("%d", len(urlParts[2])) + "\r\n\r\n"))
		conn.Write([]byte(urlParts[2]))

	case "user-agent":
		for _, header := range strings.Split(headers, "\r\n") {
			headerName, headerValue := splitByFirstOccurrence(header, ": ")

			if strings.ToLower(headerName) == "user-agent" {
				conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: " + fmt.Sprintf("%d", len(headerValue)) + "\r\n\r\n"))
				conn.Write([]byte(headerValue))
			}
		}

	default:
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}
}

func splitByFirstOccurrence(s, sep string) (string, string) {
	parts := strings.SplitN(s, sep, 2)
	return parts[0], parts[1]
}
