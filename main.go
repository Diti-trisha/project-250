package main

import (
	"fmt"
	"log"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello world")
}

func main() {
	http.HandleFunc("/", handler)
	log.Println("Server is running on port 8080")
	http.ListenAndServe(":8080", nil)
}
