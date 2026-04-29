package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Initialize the server and start listening for requests
	// This is where you would set up your routes, handlers, and any necessary middleware
	//
	// For now, simple restapi that prints the health of the system

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Server is healthy"))
	})

	quit := make(chan os.Signal, 1)
	go http.ListenAndServe(":8080", mux)

	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	// Perform any necessary cleanup before shutting down
	// This could include closing database connections, stopping background tasks, etc.

	fmt.Println("Shutting down server...")
	fmt.Println("Server stopped.")

	os.Exit(0)
}
