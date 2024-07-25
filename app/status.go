package main

import "net"

var statusCodeToText = map[int]string{
	200: "OK",
	201: "Created",
	400: "Bad Request",
	404: "Not Found",
	500: "Internal Server Error",
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
