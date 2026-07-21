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
	ServerName string `json:"server_name"`
	GameName string `json:"game_name"`
}

func main() {
	http.HandleFunc("/containers/start", handleStartServer)

	if err:= http.ListenAndServe(":8081", nil); err!= nil {
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
		http.Error(w, "Invalid JSON body. 'server_name' is required.", http.StatusBadRequest)
		return
	}

	fmt.Printf("Received instruction to spin up server for: %s\n", req.GameName)
	fmt.Printf("It will be called: %s\n", req.ServerName)

	// Context handles timeouts and cancellation signals in Go
	ctx := context.Background()

	// 1. Initialize the Docker client from your environment variables
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		fmt.Printf("Failed to connect to Docker daemon: %v\n", err)
		return
	}
	defer cli.Close()

	fmt.Println("Successfully connected to local Docker Daemon!")

	// 2. Set up the configs

	config, hostConfig, err := configSetup(req.GameName, cli, ctx)
//
	// 3. Create the container instance
	resp, err := cli.ContainerCreate(ctx, config, hostConfig, nil, nil, req.ServerName)
	if err != nil {
		fmt.Printf("Failed to create container: %v\n", err)
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
	json.NewEncoder(w).Encode(map[string]string{
		"status":       "created",
		"container_id": resp.ID,
		"server_name":  req.ServerName,
	})

}