package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/hello", hello)
	http.ListenAndServe(":8080", nil)
}

func hello(writer http.ResponseWriter, req *http.Request) {
	log.Println("/hello endpoint requested")
	writer.Write([]byte("Hello World!"))
	return
}
