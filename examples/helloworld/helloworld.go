package main

import (
	"fmt"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("new request from ", r.RemoteAddr)
		_, _ = fmt.Fprintln(w, "Hello, world!")
	})

	fmt.Println("Listening on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		panic(err)
	}
}
