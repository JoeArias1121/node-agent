package main

import (
	"context"
	"fmt"
	"os"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

func temp() {
	ctx := context.Background()

	// 1. Initialize the Docker client targeting machine's local engine
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	// 2. Define the game image we want (Using a lightweight alpine test or minecraft container)
	imageName := "docker.io/library/alpine"

	fmt.Printf("Pulling runtime image: %s...\n", imageName)
	reader, err := cli.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		panic(err)
	}
	// Copy outputs to console to see it downloading
	os.Stdout.ReadFrom(reader) 

	// 3. Configure the container specifications
	fmt.Println("Configuring game container environment...")
	resp, err := cli.ContainerCreate(ctx, 
		&container.Config{
			Image: imageName,
			Cmd:   []string{"echo", "Game Server Started Successfully!"}, // The start command
		}, 
		&container.HostConfig{}, // Define custom RAM/CPU allocations here
		nil, nil, "test-game-container")
	if err != nil {
		panic(err)
	}

	// 4. Start the container
	fmt.Printf("Launching container with ID: %s\n", resp.ID)
	err = cli.ContainerStart(ctx, resp.ID, container.StartOptions{})
	if err != nil {
		panic(err)
	}

	fmt.Println("Success! The container is running.")
}