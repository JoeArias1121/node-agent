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

// assuming by this point name and game are all good
func configSetup(game string, cli *client.Client, ctx context.Context) (*container.Config, *container.HostConfig, error) {
	if game == "minecraft" {
		return minecraftConfig(cli, ctx)
	}

	return nil, nil, fmt.Errorf("Error with config setup")
}

func minecraftConfig(cli *client.Client, ctx context.Context) (*container.Config, *container.HostConfig, error) {
	imageName := "docker.io/itzg/minecraft-server:latest"
	fmt.Printf("Pulling Minecraft image: %s (This may take a minute...)\n", imageName)
	out, err := cli.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to pull image: %v\n", err)
	}
	defer out.Close()

	// Stream the pull progress directly to your terminal window
	io.Copy(os.Stdout, out)

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
					HostPort: "25565",   // Map it to port 25565 on your laptop/server
				},
			},
		},
		Resources: container.Resources{
			Memory:   2 * 1024 * 1024 * 1024, // Sandboxed: Strict 2GB RAM upper limit
			NanoCPUs: 1000000000,            // Sandboxed: Exactly 1 CPU Core max allocation
		},
	}

	return config, hostConfig, nil
}