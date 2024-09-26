package main

import (
	"bufio"
	"context"
	"encoding/json"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

// StripANSI removes ANSI escape codes from a string
func StripANSI(input string) string {
	re := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return re.ReplaceAllString(input, "")
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// json request body
		var reqBody struct {
			ContainerName    string   `json:"containerName"`
			ContainerTag     string   `json:"containerTag"`
			ContainerCommand []string `json:"containerCommand"`
		}

		// decode the request body into reqBody
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// check if the request body contains the required parameters
		if reqBody.ContainerName == "" || len(reqBody.ContainerCommand) == 0 {
			http.Error(w, "Missing parameters: containerName or containerCommand", http.StatusBadRequest)
			return
		}

		containerName := reqBody.ContainerName
		containerTag := reqBody.ContainerTag
		containerCommand := reqBody.ContainerCommand

		out, err := CreateContainer(containerName, containerTag, containerCommand)
		// log the error if there is any
		if err != nil {
			log.Println(err)

			// return the error as the response in JSON format
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		// Split the output into separate lines and strip ANSI codes
		lines := strings.Split(out, "\n")
		for i, line := range lines {
			lines[i] = StripANSI(line)
		}

		// return the output of the container as the response in JSON format
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string][]string{"output": lines})
	})

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func CreateContainer(containerName string, containerTag string, containerCommand []string) (string, error) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return "", err
	}
	defer cli.Close()

	// Check if the image is already available locally
	images, err := cli.ImageList(ctx, image.ListOptions{})
	if err != nil {
		return "", err
	}

	if strings.HasPrefix(containerName, ":") {
		containerName = containerName[1:]
	}

	imageFound := false
	for _, image := range images {
		for _, tag := range image.RepoTags {
			if tag == containerName+":"+containerTag {
				imageFound = true
				break
			}
		}
		if imageFound {
			break
		}
	}

	// Pull the image from docker.io if it's not found locally
	if !imageFound {
		reader, err := cli.ImagePull(ctx, "docker.io/"+containerName+":"+containerTag, image.PullOptions{})
		if err != nil {
			return "", err
		}
		defer reader.Close()
		io.Copy(os.Stdout, reader)
	}

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: containerName,
		Cmd:   containerCommand,
		Tty:   false,
	}, nil, nil, nil, "")
	if err != nil {
		return "", err
	}

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", err
	}

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return "", err
		}
	case <-statusCh:
	}

	out, err := cli.ContainerLogs(ctx, resp.ID, container.LogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		return "", err
	}

	var output string
	scanner := bufio.NewScanner(out)
	for scanner.Scan() {
		line := scanner.Text()
		output += line + "\n"
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}

	// Remove the container in the background
	go func() {
		if err := cli.ContainerRemove(ctx, resp.ID, container.RemoveOptions{}); err != nil {
			log.Println("Error removing container:", err)
		}
	}()

	return output, nil
}
