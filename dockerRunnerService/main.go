package main

import (
	"encoding/json"
	"log"
	"main/ansi"
	"main/docker"
	"net/http"
	"strings"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// json request body
		var reqBody struct {
			ContainerName    string   `json:"containerName"`
			ContainerTag     string   `json:"containerTag"`
			ContainerCommand []string `json:"containerCommand"`
			Volume           []string `json:"volume"`
			Env              []string `json:"env"`
		}

		// decode the request body into reqBody
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		containerName := reqBody.ContainerName
		containerTag := reqBody.ContainerTag
		containerCommand := reqBody.ContainerCommand
		volumes := reqBody.Volume
		env := reqBody.Env

		out, err := docker.CreateContainer(containerName, containerTag, containerCommand, volumes, env)
		// log the error if there is any
		if err != nil {
			log.Println(err)

			// return the error as the response in JSON format
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		ansi.Strip(out)
		lines := strings.Split(out, "\n")

		// return the output of the container as the response in JSON format
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(lines)
	})

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
