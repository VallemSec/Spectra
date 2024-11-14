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

func CheckLocalImg(ctx context.Context, c *client.Client, iN string) (bool, error) {
	images, err := c.ImageList(ctx, image.ListOptions{})
	if err != nil {
		return false, err
	}

	for _, img := range images {
		for _, tag := range img.RepoTags {
			if tag == iN {
				return true, nil
			}
		}
	}

	return false, nil
}

func StartAndReadLogs(ctx context.Context, c *client.Client, containerName string, containerCommand, volumes, envVars []string) (*bytes.Buffer, string, error) {
	hostConfig := &container.HostConfig{
		Binds: volumes,
	}

	resp, err := c.ContainerCreate(ctx, &container.Config{
		Image: containerName,
		Cmd:   containerCommand,
		Tty:   false,
		Env:   envVars,
	}, hostConfig, nil, nil, "")
	if err != nil {
		return nil, "", err
	}

	err = c.ContainerStart(ctx, resp.ID, container.StartOptions{})
	if err != nil {
		return nil, "", err
	}

	statusCh, errCh := c.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return nil, "", err
		}
	case <-statusCh:
	}

	out, err := c.ContainerLogs(ctx, resp.ID, container.LogsOptions{ShowStdout: true, ShowStderr: true, Follow: true})
	if err != nil {
		return nil, "", err
	}

	defer out.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, out)
	if err != nil {
		return nil, "", err
	}

	return buf, resp.ID, nil
}
