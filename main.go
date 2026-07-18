package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

func main() {
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
	fmt.Println("\nMinecraft image pulled successfully!")

	// 3. Configure the container specifications
	fmt.Println("Configured Minecraft Environment settings...")

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
	resp, err := cli.ContainerCreate(ctx, config, hostConfig, nil, nil, "my-go-minecraft-server")
	if err != nil {
		fmt.Printf("Failed to create container: %v\n", err)
		return
	}

	// 5. Start the container executing
	fmt.Printf("Launching Minecraft Server Container ID: %s\n", resp.ID)
	err = cli.ContainerStart(ctx, resp.ID, container.StartOptions{})
	if err != nil {
		fmt.Printf("Failed to start container: %v\n", err)
		return
	}

	fmt.Println("\nSuccess! Your containerized Minecraft server is booting up.")
	fmt.Println("To view the active server startup logs, run: docker logs -f my-go-minecraft-server")

}