package main

import (
	"log"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("start")
		defer log.Println("end")

		longOperation()
	})

	addr := ":8080"
	log.Printf("listening on address %q", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func longOperation() {
	for i, n := 1, 10; i < n; i++ {
		time.Sleep(time.Second)

		log.Printf("%v/%v", i, n)
	}
}
