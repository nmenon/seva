/*
Wrapper for docker compose functions
*/
package main

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/melbahja/got"
)

type Containers []struct {
	ID         string `json:"ID"`
	Name       string `json:"Name"`
	Command    string `json:"Command"`
	Project    string `json:"Project"`
	Service    string `json:"Service"`
	State      string `json:"State"`
	Health     string `json:"Health"`
	ExitCode   int    `json:"ExitCode"`
	Publishers []struct {
		URL           string `json:"URL"`
		TargetPort    int    `json:"TargetPort"`
		PublishedPort int    `json:"PublishedPort"`
		Protocol      string `json:"Protocol"`
	} `json:"Publishers"`
}

func start_app(command WebSocketCommand) WebSocketCommand {
	log.Println("Starting selected app")
	cmd := exec.Command(docker_compose, "-p", "seva-launcher", "up", "-d")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Println("Failed to start selected app!")
		exit(1)
	}
	output_s := strings.TrimSpace(string(output))
	command.Response = append(command.Response, strings.Split(output_s, "\n")...)
	log.Printf("|\n%s\n", output_s)
	return command
}

func stop_app(command WebSocketCommand) WebSocketCommand {
	log.Println("Stopping selected app")
	cmd := exec.Command(docker_compose, "-p", "seva-launcher", "down", "--remove-orphans")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Println("Failed to stop selected app! (It may not be running!)")
	}
	output_s := strings.TrimSpace(string(output))
	command.Response = append(command.Response, strings.Split(output_s, "\n")...)
	log.Printf("|\n%s\n", output_s)
	return command
}

func get_app(command WebSocketCommand) WebSocketCommand {
	if _, err := os.Stat("metadata.json"); errors.Is(err, os.ErrNotExist) {
		command.Response = append(command.Response, "{}")
		return command
	}
	content, err := os.ReadFile("metadata.json")
	if err != nil {
		log.Println(err)
		exit(1)
	}
	command.Response = []string{string(content)}
	return command
}

func load_app(command WebSocketCommand) WebSocketCommand {
	name := command.Arguments[0]
	log.Println("Loading " + name + " from store")
	command = stop_app(command)

	files := []string{"metadata.json", "docker-compose.yml"}
	for _, element := range files {
		if _, err := os.Stat(element); errors.Is(err, os.ErrNotExist) {
			continue
		}
		err := os.Remove(element)
		if err != nil {
			log.Println("Failed to remove old files")
			exit(1)
		}
	}
	g := got.New()
	for _, element := range files {
		url := store_url + "/" + name + "/" + element
		log.Println("Fetching " + element + " from: " + url)
		err := g.Download(url, element)
		if err != nil {
			log.Println(err)
			exit(1)
		}
	}
	command.Response = append(command.Response, "0")
	return command
}

func is_running(command WebSocketCommand) WebSocketCommand {
	name := command.Arguments[0]
	log.Println("Checking if " + name + " is running")
	cmd := exec.Command(docker_compose, "-p", "seva-launcher", "ps", "--format", "json")
	output, err := cmd.Output()
	if err != nil {
		log.Println("Failed to check if app is running!")
		exit(1)
	}
	var containers Containers
	err = json.Unmarshal([]byte(output), &containers)
	if err != nil {
		log.Println("Failed to parse JSON from docker-compose!")
		exit(1)
	}
	for _, element := range containers {
		if element.Name == name {
			command.Response = append(command.Response, "1")
			return command
		}
	}
	command.Response = append(command.Response, "0")
	return command
}
