package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/healthcheck", healthcheck)

	fmt.Println("Server is running on port 4000")

	err := http.ListenAndServe(":4000", mux)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func healthcheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	fmt.Fprintln(w, "status: available")
	fmt.Fprintf(w, "environment: %s\n", "development")
	fmt.Fprintf(w, "version: %s\n", "1.0.0")
}
