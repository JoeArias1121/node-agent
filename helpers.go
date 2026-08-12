package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

// assuming by this point name and game are all good
func configSetup(game string, cli *client.Client, ctx context.Context, memoryGB int64, cpus float64) (*container.Config, *container.HostConfig, int, error) {
	if game == "Minecraft" {
		return minecraftConfig(cli, ctx, memoryGB, cpus)
	}

	return nil, nil, 0, fmt.Errorf("Error with config setup")
}

func minecraftConfig(cli *client.Client, ctx context.Context, memoryGB int64, cpus float64) (*container.Config, *container.HostConfig, int, error) {
	imageName := "docker.io/itzg/minecraft-server:latest"
	fmt.Printf("Pulling Minecraft image: %s (This may take a minute...)\n", imageName)
	out, err := cli.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return nil, nil, 0, fmt.Errorf("Failed to pull image: %v\n", err)
	}
	defer out.Close()

	// Stream the pull progress directly to your terminal window
	io.Copy(os.Stdout, out)

	// Dynamically find a free port on the host
	port, err := getFreePort()
	if err != nil {
		return nil, nil, 0, fmt.Errorf("Failed to allocate dynamic port: %v\n", err)
	}

	// Define the container environment variables and internal settings
	config := &container.Config{
		Image: imageName,
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
					HostPort: fmt.Sprintf("%d", port),   // Map it to dynamic port on host
				},
			},
		},
		Resources: container.Resources{
			Memory:   memoryGB * 1024 * 1024 * 1024, // Configurable RAM limit
			NanoCPUs: int64(cpus * 1000000000),      // Configurable CPU limit
		},
	}

	return config, hostConfig, port, nil
}

// getFreePort asks the OS to allocate a random open port on the loopback interface, grabs the number, and releases it
func getFreePort() (int, error) {
    // Binding to ":0" forces the OS to pick an unallocated ephemeral port automatically
    listener, err := net.Listen("tcp", ":0")
    if err != nil {
        return 0, err
    }
    defer listener.Close()

    // Extract the port number that the OS assigned to us
    localAddr := listener.Addr().(*net.TCPAddr)
    return localAddr.Port, nil
}