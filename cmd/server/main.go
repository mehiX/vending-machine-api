package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/mehix/vending-machine-api/pkg/app"
)

var addr string

func init() {
	flag.StringVar(&addr, "l", "localhost:7777", "Listen address for the server")
}

// @title 			Vending Machine API
// @description		API for a vending machine, allowing users with a “seller” role to add, update or remove products, while users with a “buyer” role can deposit coins into the machine and make purchases
// @version 		1.0

// @contact.name	Mihai O.
// @contact.email	mihai@devops-experts.me

// @securitydefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

// @BasePath		/
// @schemes			http
func main() {

	flag.Parse()

	srvr := app.NewApp(addr, nil).HttpServer()

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
