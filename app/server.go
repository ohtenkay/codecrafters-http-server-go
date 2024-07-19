package main

import (
	"bytes"
	"fmt"
	"net"
	"os"
)

func main() {
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

	b := make([]byte, 1024)
	conn.Read(b)
	parts := bytes.Split(b, []byte(" "))

	if string(parts[1]) == "/" {
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	}

	urlParts := bytes.Split(parts[1], []byte("/"))
	if string(urlParts[1]) == "echo" {
		length := len(urlParts[2])

		conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: " + fmt.Sprintf("%d", length) + "\r\n\r\n"))
		conn.Write(urlParts[2])
	}

	conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
}
