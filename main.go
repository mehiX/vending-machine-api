package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"
)

var addr string

func init() {
	flag.StringVar(&addr, "l", "localhost:7777", "Listen address for the server")
}

func main() {

	flag.Parse()

	srvr := NewApp(addr, nil).HttpServer()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		fmt.Printf("Start listening on %s\n", srvr.Addr)
		if err := srvr.ListenAndServe(); err != nil {
			fmt.Println(err)
			select {
			case c <- os.Interrupt:
			default:
			}
		}
	}()

	<-c
	fmt.Println("Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srvr.Shutdown(ctx); err != nil {
		fmt.Println(err)
	}

	fmt.Println("Done")
}
