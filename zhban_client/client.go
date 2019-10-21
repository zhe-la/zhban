// Package main implements a client for Greeter service.
package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/poloten4ik100/zhban/api"

	"google.golang.org/grpc"
)

func main() {
	var key, address, url string
	flag.StringVar(&key, "k", "", "Security key. If not set, the key will be empty string")
	flag.StringVar(&address, "addr", "", "Address to connect to gRPC server (example: 127.0.0.1)")
	flag.StringVar(&url, "url", "", "request url")
	flag.Parse()

	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		fmt.Println("did not connect", err)
	}
	defer conn.Close()
	c := api.NewZhbanClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	r, err := c.GetDataKey(ctx, &api.DataRequestKey{Key: key, Url: url})
	if err != nil {
		fmt.Println("Request error", err)
	}
	fmt.Println("RESULT:", r)
}
