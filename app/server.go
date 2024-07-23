package main

import (
	"fmt"
	"net"
	"os"
	"strings"
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

func handlecConnection(conn net.Conn, dirname string) {
	defer conn.Close()

	b := make([]byte, 1024)
	conn.Read(b)

	lineHeaders, body := splitByFirstOccurrence(string(b), "\r\n\r\n")
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
	case "files":
		switch lineParts[0] {
		case "GET":
			file, err := os.Open(dirname + urlParts[2])
			if err != nil {
				conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
				return
			}

			fileInfo, _ := file.Stat()
			fileSize := fileInfo.Size()
			fileContent := make([]byte, fileSize)
			file.Read(fileContent)

			conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: " + fmt.Sprintf("%d", fileSize) + "\r\n\r\n"))
			conn.Write(fileContent)
		case "POST":
			file, err := os.Create(dirname + urlParts[2])
			if err != nil {
				conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
				return
			}

			file.Write([]byte(strings.TrimRight(body, "\x00")))
			file.Close()

			conn.Write([]byte("HTTP/1.1 201 Created\r\n\r\n"))
		}
	default:
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}
}

func splitByFirstOccurrence(s, sep string) (string, string) {
	parts := strings.SplitN(s, sep, 2)
	return parts[0], parts[1]
}
