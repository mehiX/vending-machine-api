package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"
)

var addr string

type app struct {
	Addr   string
	Router http.Handler
	Db     *sql.DB
}

func (a *app) HttpServer() http.Server {
	return http.Server{
		Addr:    a.Addr,
		Handler: a.Router,
	}
}

func init() {
	flag.StringVar(&addr, "l", "localhost:7777", "Listen address for the server")
}

func main() {

	flag.Parse()

	vm := &app{
		Addr:   addr,
		Router: router(),
	}

	srvr := vm.HttpServer()

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
