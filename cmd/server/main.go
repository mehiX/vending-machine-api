//go:generate swag init -d "." -o "../../docs" --parseDependency --parseInternal --parseDepth 1
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/joho/godotenv"
	"github.com/mehiX/vending-machine-api/internal/app"
)

var addr string
var envFile string

func init() {
	flag.StringVar(&addr, "l", "localhost:7777", "Listen address for the server")
	flag.StringVar(&envFile, "e", ".env", "File with environment variables")

	flag.Parse()

	if envFile != "" {
		if err := godotenv.Load(envFile); err != nil {
			fmt.Printf("ENV not loaded from '%s'. Error: %s\n", envFile, err.Error())
		}
	}
}

// @title  Vending Machine API
// @description API for a vending machine, allowing users with a “seller” role to add, update or remove products, while users with a “buyer” role can deposit coins into the machine and make purchases
// @version 1.0.2

// @contact.name Mihai O.
// @contact.email mihai@devops-experts.me

// @securitydefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

// @host localhost:7777
// @BasePath /
// @schemes http
func main() {

	vm := app.NewApp(addr, nil)

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

	done, stopDB := context.WithCancel(context.Background())
	go ConnectDB(done, vm, os.Getenv("MYSQL_CONN_STR"), 5*time.Second)

	<-c
	fmt.Println("Shutting down...")

	// disconnect the database
	stopDB()

	// shutdown http server
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srvr.Shutdown(ctx); err != nil {
		fmt.Println(err)
	}

	fmt.Println("Done")
}

// ConnectDB tries to establish a database connection.
// Retries periodically to check that the connection is still available.
// Should be run in a separate goroutine.
func ConnectDB(done context.Context, myApp *app.App, connStr string, pingDelay time.Duration) {

	if myApp == nil {
		return
	}

	test := func(db *sql.DB) error {
		ctx, cancel := context.WithTimeout(done, 2*time.Second)
		defer cancel()
		return db.PingContext(ctx)
	}

	// don't fill the logs if connection is OK
	var printConnOK bool = true

	tkr := time.NewTicker(pingDelay)
	for {
		select {
		case <-done.Done():
			if myApp.Db != nil {
				myApp.Db.Close()
			}
			fmt.Println("DB connection closed")
			return
		case <-tkr.C:
			if myApp.Db == nil {
				// try to connect
				fmt.Println("DB: connecting...")
				db, err := sql.Open("mysql", connStr)
				if err != nil {
					fmt.Printf("DB: %v\n", err.Error())
				} else {
					db.SetConnMaxLifetime(0)
					db.SetMaxIdleConns(50)
					db.SetMaxOpenConns(50)

					if err := test(db); err == nil {
						myApp.Db = db
					}
				}
			} else {
				// check if server still available
				if err := test(myApp.Db); err != nil {
					fmt.Printf("DB: Ping %v\n", err.Error())
					myApp.Db = nil
					printConnOK = true
				} else {
					if printConnOK {
						fmt.Println("DB: connection OK")
						printConnOK = false
					}
				}
			}
		}
	}
}
