package docker

import (
	"bytes"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	_ "github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"
	"io"
)

func CheckLocalImg(ctx context.Context, client *client.Client, imageName string) (bool, error) {
	images, err := client.ImageList(ctx, image.ListOptions{})
	if err != nil {
		return false, err
	}

	for _, img := range images {
		for _, tag := range img.RepoTags {
			if tag == imageName {
				return true, nil
			}
		}
	}

	return false, nil
}

func StartAndReadLogs(ctx context.Context, client *client.Client, containerName string, containerCommand []string) (string, error) {
	resp, err := client.ContainerCreate(ctx, &container.Config{
		Image: containerName,
		Cmd:   containerCommand,
		Tty:   false,
	}, nil, nil, nil, "")
	if err != nil {
		return "", err
	}

	err = client.ContainerStart(ctx, resp.ID, container.StartOptions{})
	if err != nil {
		return "", err
	}

	statusCh, errCh := client.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return "", err
		}
	case <-statusCh:
	}

	out, err := client.ContainerLogs(ctx, resp.ID, container.LogsOptions{ShowStdout: true, ShowStderr: true, Follow: true})
	if err != nil {
		return "", err
	}

	defer out.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, out)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
