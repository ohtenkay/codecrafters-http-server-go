package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"net"
	"os"
	"strings"
)

var statusCodeToText = map[int]string{
	200: "OK",
	201: "Created",
	400: "Bad Request",
	404: "Not Found",
	500: "Internal Server Error",
}

type request struct {
	method    string
	pathParts [][]byte
	version   string
	headers   map[string]string
	body      []byte
}

func newRequest(buff []byte) (*request, error) {
	lineHeaders, body, found := bytes.Cut(buff, []byte("\r\n\r\n"))
	if !found {
		return nil, fmt.Errorf("bad request")
	}

	line, headers, found := bytes.Cut(lineHeaders, []byte("\r\n"))
	if !found {
		return nil, fmt.Errorf("bad request")
	}

	method, path, version, err := splitLine(line)
	if err != nil {
		return nil, err
	}

	pathParts := bytes.Split(path, []byte("/"))
	if len(pathParts) < 2 {
		return nil, fmt.Errorf("invlid path")
	}

	request := &request{
		method:    string(method),
		pathParts: pathParts,
		version:   string(version),
		headers:   make(map[string]string),
		body:      body,
	}

	for _, header := range bytes.Split(headers, []byte("\r\n")) {
		headerName, headerValue, found := bytes.Cut(header, []byte(": "))
		if !found {
			return nil, fmt.Errorf("invalid header" + string(header))
		}

		request.headers[string(headerName)] = string(headerValue)
	}

	return request, nil
}

type response struct {
	request    *request
	statusCode int
	headers    map[string]string
	body       []byte
}

func (r *response) write(conn net.Conn) {
	if r.request != nil {
		for key, value := range r.request.headers {
			if strings.ToLower(key) == "accept-encoding" {
				for _, encoding := range strings.Split(value, ",") {
					if strings.TrimSpace(encoding) == "gzip" {
						r.headers["Content-Encoding"] = "gzip"

						compressedBody, err := compress(r.body)
						if err != nil {
							respondInternalServerError(conn)
							return
						}

						r.body = compressedBody
						r.headers["Content-Length"] = fmt.Sprintf("%d", len(r.body))
						break
					}
				}

				break
			}
		}
	}

	conn.Write([]byte("HTTP/1.1 " + fmt.Sprintf("%d", r.statusCode) + " " + statusCodeToText[r.statusCode] + "\r\n"))
	for key, value := range r.headers {
		conn.Write([]byte(key + ": " + value + "\r\n"))
	}
	conn.Write([]byte("\r\n"))
	conn.Write(r.body)
}

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

func respondBadRequest(conn net.Conn) {
	response := &response{
		statusCode: 400,
	}
	response.write(conn)
}

func respondNotFound(conn net.Conn) {
	response := &response{
		statusCode: 404,
	}
	response.write(conn)
}

func respondInternalServerError(conn net.Conn) {
	response := &response{
		statusCode: 500,
	}
	response.write(conn)
}

func splitLine(line []byte) ([]byte, []byte, []byte, error) {
	parts := bytes.Split(line, []byte(" "))
	if len(parts) != 3 {
		return nil, nil, nil, fmt.Errorf("invalid line")
	}

	return parts[0], parts[1], parts[2], nil
}

func compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer

	gzipWriter := gzip.NewWriter(&buf)

	_, err := gzipWriter.Write(data)
	if err != nil {
		return nil, err
	}

	if err := gzipWriter.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
