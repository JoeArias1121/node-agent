package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type StartRequest struct {
	ServerName string  `json:"server_name"`
	GameName   string  `json:"game_name"`
	MemoryGB   int64   `json:"memory_gb"` // Optional: RAM limit in GB (e.g. 2)
	CPUs       float64 `json:"cpus"`      // Optional: CPU allocation in cores (e.g. 1.0)
}

type StopRequest struct {
	ContainerID string `json:"container_id"`
}

func main() {
	http.HandleFunc("/containers/start", handleStartServer)
	http.HandleFunc("/containers/stop", handleStopServer)

	fmt.Println("Agent listening on port :8081...")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		fmt.Printf("Agent failed to start: %v\n", err)
	}
}

func handleStartServer(w http.ResponseWriter, r *http.Request) {
	// Restrict this endpoint to HTTP POST requests only
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed", http.StatusMethodNotAllowed)
		return
	}
	// Parse the incoming JSON payload from the request body
	var req StartRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.ServerName == "" || req.GameName == "" {
		http.Error(w, "Invalid JSON body. 'server_name' and 'game_name' are required.", http.StatusBadRequest)
		return
	}

	// Default to 2GB RAM and 1.0 CPU if not specified
	if req.MemoryGB <= 0 {
		req.MemoryGB = 2
	}
	if req.CPUs <= 0 {
		req.CPUs = 1.0
	}

	fmt.Printf("Received instruction to spin up server for: %s (Size: %d GB RAM, %.1f CPU)\n", req.GameName, req.MemoryGB, req.CPUs)
	fmt.Printf("It will be called: %s\n", req.ServerName)

	// Context handles timeouts and cancellation signals in Go
	ctx := context.Background()

	// 1. Initialize the Docker client from your environment variables
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		fmt.Printf("Failed to connect to Docker daemon: %v\n", err)
		http.Error(w, fmt.Sprintf("Failed to connect to Docker daemon: %v", err), http.StatusInternalServerError)
		return
	}
	defer cli.Close()

	fmt.Println("Successfully connected to local Docker Daemon!")

	// 2. Set up the configs
	config, hostConfig, port, err := configSetup(req.GameName, cli, ctx, req.MemoryGB, req.CPUs)
	if err != nil {
		fmt.Printf("Failed to setup container config: %v\n", err)
		http.Error(w, fmt.Sprintf("Configuration error: %v", err), http.StatusBadRequest)
		return
	}

	// 3. Create the container instance
	resp, err := cli.ContainerCreate(ctx, config, hostConfig, nil, nil, req.ServerName)
	if err != nil {
		fmt.Printf("Failed to create container: %v\n", err)
		http.Error(w, fmt.Sprintf("Failed to create container: %v", err), http.StatusInternalServerError)
		return
	}

	// 4. Start the container executing
	err = cli.ContainerStart(ctx, resp.ID, container.StartOptions{})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to start container: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":       "created",
		"container_id": resp.ID,
		"server_name":  req.ServerName,
		"port":         port,
	})
}

func handleStopServer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed", http.StatusMethodNotAllowed)
		return
	}

	var req StopRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.ContainerID == "" {
		fmt.Printf("Invalid JSON body %v\n", err)
		http.Error(w, "Invalid JSON body. 'container_id' is required", http.StatusBadRequest)
		return
	}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		fmt.Printf("Failed to connect to Docker daemon: %v\n", err)
		http.Error(w, fmt.Sprintf("Failed to connect to Docker daemon: %v", err), http.StatusInternalServerError)
		return
	}
	defer cli.Close()

	ctx := context.Background()
	err = cli.ContainerStop(ctx, req.ContainerID, container.StopOptions{})
	if err != nil {
		fmt.Printf("Failed to stop container: %v\n", err)
		http.Error(w, fmt.Sprintf("Failed to stop container: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":       "stopped",
		"container_id": req.ContainerID,
	})
}