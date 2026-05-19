package main

import (
	"log"
	"net/http"

	"github.com/hornflakes/go-runs-rest-walk/internal/server"
)

func main() {
	http.HandleFunc("/", server.HandleEcho)
	log.Fatal(http.ListenAndServe(":37373", nil))
}
