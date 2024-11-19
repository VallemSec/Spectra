package docker

import (
	"bytes"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	_ "github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
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

func StartAndReadLogs(ctx context.Context, c *client.Client, containerName string, containerCommand, volumes, envVars []string, tty bool) (*bytes.Buffer, *bytes.Buffer, string, error) {
	hostConfig := &container.HostConfig{
		Binds: volumes,
	}

	resp, err := c.ContainerCreate(ctx, &container.Config{
		Image: containerName,
		Cmd:   containerCommand,
		Tty:   tty,
		Env:   envVars,
	}, hostConfig, nil, nil, "")
	if err != nil {
		return nil, nil, "", err
	}

	err = c.ContainerStart(ctx, resp.ID, container.StartOptions{})
	if err != nil {
		return nil, nil, "", err
	}

	statusCh, errCh := c.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return nil, nil, "", err
		}
	case <-statusCh:
	}

	out, err := c.ContainerLogs(ctx, resp.ID, container.LogsOptions{ShowStdout: true, ShowStderr: true, Follow: true})
	if err != nil {
		return nil, nil, "", err
	}
	defer out.Close()

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	if !tty {
		_, err = stdcopy.StdCopy(stdout, stderr, out)
		if err != nil {
			return nil, nil, "", err
		}
	} else {
		_, err = io.Copy(stdout, out)
		if err != nil {
			return nil, nil, "", err
		}
	}

	return stdout, stderr, resp.ID, nil
}
