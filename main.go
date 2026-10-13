package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type helloResponse struct {
	Message string `json:"message"`
}

func healthHandler(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, err := fmt.Fprintln(writer, "ok")
	if err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

func helloHandler(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	response := helloResponse{
		Message: "Hello, world!",
	}
	err := json.NewEncoder(writer).Encode(response)
	if err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

func newRouter() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("GET /hello", helloHandler)
	return mux

}

func main() {

	log.Println("Go Serve is starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe("127.0.0.1:8080", newRouter()))

}
