package main

import (
	"bufio"
	"fmt"
	"net"

	"github.com/dominikbraun/gohttp"
)

func main() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			panic(err)
		}
		go func() {
			connReader := bufio.NewReader(conn)
			request, _ := gohttp.ParseRequest(connReader)
			fmt.Println(request)
		}()
	}
}
