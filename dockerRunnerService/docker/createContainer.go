package docker

import (
	"bufio"
	"context"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"io"
	"log"
	"os"
)

func CreateContainer(containerName, containerTag string, containerCommand, volumes, networks, env []string, tty bool) (string, string, error) {
	fmt.Println("Creating container", containerName+":"+containerTag)

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return "", "", err
	}
	defer cli.Close()

	imageName := containerName + ":" + containerTag
	imageFound, err := CheckLocalImg(ctx, cli, imageName)
	if err != nil {
		return "", "", err
	}

	fmt.Println("Image found: ", imageFound, imageName)

	// Pull the image from docker.io if it's not found locally
	if !imageFound {
		fmt.Println("Pulling image", imageName)
		reader, err := cli.ImagePull(ctx, "docker.io/"+imageName, image.PullOptions{})
		if err != nil {
			return "", "", fmt.Errorf("failed to pull image: %w", err)
		}
		defer reader.Close()
		io.Copy(os.Stdout, reader)
	}

	stdoutBuff, stderrBuff, containerId, err := StartAndReadLogs(ctx, cli, imageName, containerCommand, volumes, networks, env, tty)
	if err != nil {
		return "", "", err
	}

	if stdoutBuff == nil {
		return "", "", fmt.Errorf("failed to get logs from container")
	}

	var stdout string
	stdoutScanner := bufio.NewScanner(stdoutBuff)
	for stdoutScanner.Scan() {
		line := stdoutScanner.Text()
		stdout += line + "\n"
	}
	if err := stdoutScanner.Err(); err != nil {
		return "", "", err
	}

	var stderr string
	stderrScanner := bufio.NewScanner(stderrBuff)
	for stderrScanner.Scan() {
		line := stderrScanner.Text()
		stderr += line + "\n"
	}
	if err := stderrScanner.Err(); err != nil {
		return "", "", err
	}

	// Remove the container in the background
	go func() {
		if err := cli.ContainerRemove(ctx, containerId, container.RemoveOptions{}); err != nil {
			log.Println("Error removing container:", err)
		}
	}()

	return stdout, stderr, nil
}
