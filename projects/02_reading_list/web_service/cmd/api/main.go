package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/api/v1/healthcheck", healthcheck)

	fmt.Println("Server is running on port 4000")

	err := http.ListenAndServe(":4000", nil)

	if err != nil {
		fmt.Println(err)
	}
}

func healthcheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "status: available")
	fmt.Fprintf(w, "environment: %s\n", "development")
	fmt.Fprintf(w, "version: %s\n", "1.0.0")
}
