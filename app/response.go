package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"net"
	"strings"
)

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
