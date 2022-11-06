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

func start_app() string {
	log.Println("Starting selected app")
	cmd := exec.Command(docker_compose, "-p", "seva-launcher", "up", "-d")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Println("Failed to start selected app!")
		exit(1)
	}
	output_s := strings.TrimSpace(string(output))
	log.Printf("|\n%s\n", output_s)
	return output_s
}

func stop_app() string {
	log.Println("Stopping selected app")
	cmd := exec.Command(docker_compose, "-p", "seva-launcher", "down", "--remove-orphans")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Println("Failed to stop selected app! (It may not be running!)")
	}
	output_s := strings.TrimSpace(string(output))
	log.Printf("|\n%s\n", output_s)
	return output_s
}

func get_app() string {
	if _, err := os.Stat("metadata.json"); errors.Is(err, os.ErrNotExist) {
		return "{}"
	}
	content, err := os.ReadFile("metadata.json")
	if err != nil {
		log.Println(err)
		exit(1)
	}
	return string(content)
}

func load_app(name string) string {
	log.Println("Loading " + name + " from store")
	stop_app()

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
	return string("0")
}

func is_running(name string) string {
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
			return string("1")
		}
	}
	return string("0")
}
