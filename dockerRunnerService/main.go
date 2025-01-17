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
			Networks         []string `json:"network"`
			Env              []string `json:"env"`
			Tty              bool     `json:"tty"`
		}

		type returnBody struct {
			Stdout []string `json:"stdout"`
			Stderr []string `json:"stderr"`
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
		networks := reqBody.Networks
		env := reqBody.Env
		tty := reqBody.Tty

		stdout, stderr, err := docker.CreateContainer(containerName, containerTag, containerCommand, volumes, networks, env, tty)
		// log the error if there is any
		if err != nil {
			log.Println(err)

			// return the error as the response in JSON format
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		ansi.Strip(stdout)
		stdoutLines := strings.Split(stdout, "\n")
		stderrLines := strings.Split(stderr, "\n")
		//rBody := returnBody{stdout: stdoutLines, stderr: stderrLines}
		rBody := new(returnBody)
		rBody.Stdout = stdoutLines
		rBody.Stderr = stderrLines

		// return the output of the container as the response in JSON format
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rBody)
	})

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
