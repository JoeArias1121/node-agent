package main

import (
	"fmt"
	"flag"
	"net/http"
	"encoding/json"
	"node-agent/sysstats"
)

type StopRequest struct {
	ContainerID string `json:"container_id"`
}

func main() {
	portPtr := flag.Int("port", 8090, "Default port number")

	flag.Parse()

	port := fmt.Sprintf(":%v", *portPtr)

	http.HandleFunc("/api/v1/specs", handleGetSpecs)

	fmt.Println("Agent listening on port ", port, "...")
	if err := http.ListenAndServe(port, nil); err != nil {
		fmt.Printf("Agent failed to start: %v\n", err)
	}

}

func handleGetSpecs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, fmt.Sprintln("Wrong http method"), http.StatusMethodNotAllowed)
		return
	}
	specs, err := sysstats.GetSpecs()
	if err != nil {
		http.Error(w, fmt.Sprintf("Issue getting the device specs: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(specs); err != nil {
		http.Error(w, "failed to encode specs", http.StatusInternalServerError)
	}
}
