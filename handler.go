package main

import (
	"net/http"
)

type HttpHandler struct {
	Containerizer *Containerizer
	Discoverer *Discoverer
}

func (h HttpHandler) HandleTop(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello"))
}

