package main

import (
	"net/http"
)

func main() {
	err := http.ListenAndServe("localhost:5008", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := "HTTP upstream Ready!\n\n"

		w.WriteHeader(200)
		w.Header().Add("Content-Type", "text/plain")
		w.Write([]byte(body))
		r.Header.Write(w)
	}))

	panic(err)
}
