package main

import (
	"log"
	"encoding/json"
	"net/http"
	"github.com/gorilla/mux"
	"io/ioutil"
)

type HttpHandler struct {
	Containerizer Containerizer
	Discoverer Discoverer
}

func (h HttpHandler) HandleError(err error, w http.ResponseWriter) bool {
	if (err != nil) {
		log.Println(err)
		http.Error(w, http.StatusText(500), 500)
		return true
	}
	return false
}

func (h HttpHandler) LogRequest(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL)
}

func (h HttpHandler) HandleTop(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(w, r)
	w.Write([]byte("Hello Dockertie"))
}

func (h HttpHandler) HandleHosts(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(w, r)

	hosts, err := h.Discoverer.GetHosts(nil)
	if (h.HandleError(err, w)) {
		return
	}

	b, err := json.Marshal(hosts)
	if (h.HandleError(err, w)) {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func (h HttpHandler) HandleHostContainers(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(w, r)
	vars := mux.Vars(r)

	ids := []string{vars["id"]}
	hosts, err := h.Discoverer.GetHosts(ids)
	if (h.HandleError(err, w)) {
		return
	}

	host := hosts[0]

	containers, err := h.Containerizer.GetContainersOnHost(host)
	if (h.HandleError(err, w)) {
		return
	}

	b, err := json.Marshal(containers)
	if (h.HandleError(err, w)) {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func (h HttpHandler) HandleContainers(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(w, r)

	switch r.Method {
	case "GET":
		h.HandleContainersGet(w, r)
	case "POST":
		h.HandleContainersPost(w, r)
	default:
		http.Error(w, http.StatusText(404), 404)
	}
}

func (h HttpHandler) HandleContainersGet(w http.ResponseWriter, r *http.Request) {
	hosts, err := h.Discoverer.GetHosts(nil)
	if (h.HandleError(err, w)) {
		return
	}

	containers, err := h.Containerizer.GetContainersOnHosts(hosts)
	if (h.HandleError(err, w)) {
		return
	}

	b, err := json.Marshal(containers)
	if (h.HandleError(err, w)) {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func (h HttpHandler) HandleContainersPost(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	if (h.HandleError(err, w)) {
		return
	}
	r.Body.Close()

	var config ContainerConfig
	json.Unmarshal(b, &config)

	hosts, err := h.Discoverer.GetHosts(nil)
	if (h.HandleError(err, w)) {
		return
	}

	host, err := h.Containerizer.FindAvailableHost(hosts)
	if (h.HandleError(err, w)) {
		return
	}

	container, err := h.Containerizer.RunContainer(*host, config)
	if (h.HandleError(err, w)) {
		return
	}

	b, err = json.Marshal(container)
	if (h.HandleError(err, w)) {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

