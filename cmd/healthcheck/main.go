package main

import (
	"log"
	"net/http"
	"os"
)

func main() {

	if len(os.Args) < 2 {
		log.Fatal("Missing status endpoint URL")
	}

	statusEndPoint := os.Args[1]

	r, err := http.Get(statusEndPoint)
	if err != nil {
		log.Fatal(err)
	}

	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		log.Fatal(r.Status)
	}

}
