package main

import (
	"go-server/handlers"
	"log"
	"net/http"
)

func main() {

	http.HandleFunc("/hello", handlers.HelloHandler)
	http.HandleFunc("/fap", handlers.FapHandler)

	log.Println("Бот запустился")

	err := http.ListenAndServe(":8080", nil)

	if err != nil {
		log.Fatal(err)
	}
}
