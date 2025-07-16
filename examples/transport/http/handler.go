package main

import (
	"fmt"
	"net/http"
)

func Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		if _, err := w.Write([]byte("hello world")); err != nil {
			fmt.Println(err)
		}
	})
}
