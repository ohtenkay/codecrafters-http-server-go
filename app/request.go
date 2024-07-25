package main

import (
	"bytes"
	"fmt"
)

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

func splitLine(line []byte) ([]byte, []byte, []byte, error) {
	parts := bytes.Split(line, []byte(" "))
	if len(parts) != 3 {
		return nil, nil, nil, fmt.Errorf("invalid line")
	}

	return parts[0], parts[1], parts[2], nil
}
