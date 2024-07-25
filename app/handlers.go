package main

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"strings"
)

func handlecConnection(conn net.Conn, dirname string) {
	defer conn.Close()

	buff := make([]byte, 1024)
	conn.Read(buff)

	request, err := newRequest(buff)
	if err != nil {
		respondBadRequest(conn)
		return
	}

	switch string(request.pathParts[1]) {
	case "":
		handleRoot(request, conn)
	case "echo":
		handleEcho(request, conn)
	case "user-agent":
		handleUserAgent(request, conn)
	case "files":
		handleFiles(request, conn, dirname)
	default:
		respondNotFound(conn)
	}
}

func handleRoot(request *request, conn net.Conn) {
	response := &response{
		request:    request,
		statusCode: 200,
	}
	response.write(conn)
}

func handleEcho(request *request, conn net.Conn) {
	response := &response{
		request:    request,
		statusCode: 200,
		headers: map[string]string{
			"Content-Type":   "text/plain",
			"Content-Length": fmt.Sprintf("%d", len(request.pathParts[2])),
		},
		body: request.pathParts[2],
	}
	response.write(conn)
}

func handleUserAgent(request *request, conn net.Conn) {
	for key, value := range request.headers {
		if strings.ToLower(key) == "user-agent" {
			response := &response{
				request:    request,
				statusCode: 200,
				headers: map[string]string{
					"Content-Type":   "text/plain",
					"Content-Length": fmt.Sprintf("%d", len(value)),
				},
				body: []byte(value),
			}
			response.write(conn)
			return
		}
	}

	respondNotFound(conn)
}

func handleFiles(request *request, conn net.Conn, dirname string) {
	switch request.method {
	case "GET":
		file, err := os.Open(dirname + string(request.pathParts[2]))
		if err != nil {
			respondNotFound(conn)
			return
		}

		fileInfo, _ := file.Stat()
		fileSize := fileInfo.Size()
		fileContent := make([]byte, fileSize)
		file.Read(fileContent)

		response := &response{
			request:    request,
			statusCode: 200,
			headers: map[string]string{
				"Content-Type":   "application/octet-stream",
				"Content-Length": fmt.Sprintf("%d", fileSize),
			},
			body: fileContent,
		}
		response.write(conn)
	case "POST":
		file, err := os.Create(dirname + string(request.pathParts[2]))
		if err != nil {
			respondInternalServerError(conn)
			return
		}

		file.Write(bytes.TrimRight(request.body, "\x00"))
		file.Close()

		response := &response{
			request:    request,
			statusCode: 201,
		}
		response.write(conn)
	}
}
