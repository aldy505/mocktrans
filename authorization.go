package main

import (
	"encoding/base64"
	"net/http"
)

func (d *Dependencies) Authorization(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for Authorization
		if r.Header.Get("Authorization") != "Basic "+base64.StdEncoding.EncodeToString([]byte(d.ServerKey+":")) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		h.ServeHTTP(w, r)
	})
}
