package main

import (
	"fmt"
	"log"
	"net/http"
)

var calls []string = make([]string, 0)
var stats map[string]int = make(map[string]int)

func main() {
	http.HandleFunc("/hello", helloHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")

	if name == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if _, ok := stats[name]; !ok {
		stats[name] = 0
	}
	stats[name]++

	calls = append(calls, name)

	fmt.Printf("calls: %#v\n", calls)
	fmt.Printf("stats: %#v\n\n", stats)

	fmt.Fprintf(w, "Hello, %s", name)
}
