package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type StartRequest struct {
	ServerName string `json:"server_name"`
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
	if err != nil || req.ServerName == "" {
		http.Error(w, "Invalid JSON body. 'server_name' is required.", http.StatusBadRequest)
		return
	}

	fmt.Printf("Received instruction to spin up server: %s\n", req.ServerName)

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

	// 2. Define the image name
	minecraftImage := "docker.io/itzg/minecraft-server:latest"

	fmt.Printf("Pulling Minecraft image: %s (This may take a minute...)\n", minecraftImage)
	out, err := cli.ImagePull(ctx, minecraftImage, image.PullOptions{})
	if err != nil {
		fmt.Printf("Failed to pull image: %v\n", err)
		return
	}
	defer out.Close()

	// Stream the pull progress directly to your terminal window
	io.Copy(os.Stdout, out)

	// Define the container environment variables and internal settings
	config := &container.Config{
		Image: minecraftImage,
		Env: []string{
			"EULA=TRUE",         // Crucial: Minecraft will crash on startup without this
			"TYPE=VANILLA",      // Spawns a standard vanilla server
			"VERSION=LATEST",    // Uses the latest stable Minecraft release
		},
		ExposedPorts: nat.PortSet{
			"25565/tcp": struct{}{}, // Expose the internal default Minecraft port
		},
	}

	// Define the host machine resources and networking boundaries
	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			"25565/tcp": []nat.PortBinding{
				{
					HostIP:   "0.0.0.0", // Allow anyone to connect to your host machine IP
					HostPort: "25565",   // Map it to port 25565 on your laptop/server
				},
			},
		},
		Resources: container.Resources{
			Memory:   2 * 1024 * 1024 * 1024, // Sandboxed: Strict 2GB RAM upper limit
			NanoCPUs: 1000000000,            // Sandboxed: Exactly 1 CPU Core max allocation
		},
	}

	// 4. Create the container instance
	resp, err := cli.ContainerCreate(ctx, config, hostConfig, nil, nil, req.ServerName)
	if err != nil {
		fmt.Printf("Failed to create container: %v\n", err)
		return
	}

	// 5. Start the container executing
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