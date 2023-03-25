/*
Wrapper for docker compose functions
*/
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/melbahja/got"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

type ProxySettings struct {
	HTTP   string `json:"http_proxy"`
	HTTPS  string `json:"https_proxy"`
	FTP    string `json:"ftp_proxy"`
	NO     string `json:"no_proxy"`
}

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

func update_sysconfig(proxy_settings ProxySettings) {
	format := `
export https_proxy="%s"
export http_proxy="%s"
export ftp_proxy="%s"
export no_proxy="%s"
`
	sysconfig_proxy := fmt.Sprintf(
		format,
		proxy_settings.HTTPS,
		proxy_settings.HTTP,
		proxy_settings.FTP,
		proxy_settings.NO,
	)

	// Write the proxy setting
	err := ioutil.WriteFile("/etc/sysconfig/docker", []byte(sysconfig_proxy), 0644)
	if err != nil {
		log.Println(err)
		exit(1)
	}

	// Restart the Docker daemon after setting up the proxy
	cmd := exec.Command("service", "docker", "restart")
	err = cmd.Run()
	if err != nil {
		log.Println(err)
		exit(1)
	}
}

func update_systemd(proxy_settings ProxySettings) {
	format := `
[Service]
Environment=https_proxy="%s"
Environment=http_proxy="%s"
Environment=ftp_proxy="%s"
Environment=no_proxy="%s"
`
	systemd_proxy := fmt.Sprintf(
		format,
		proxy_settings.HTTPS,
		proxy_settings.HTTP,
		proxy_settings.FTP,
		proxy_settings.NO,
	)

	// Create /etc/systemd/system/docker.service.d directory structure
	if err := os.MkdirAll("/etc/systemd/system/docker.service.d", os.ModePerm); err != nil {
		log.Println(err)
		exit(1)
	}

	// Write the proxy setting
	err_ := ioutil.WriteFile("/etc/systemd/system/docker.service.d/http-proxy.conf", []byte(systemd_proxy), 0644)
	if err_ != nil {
		log.Println(err_)
		exit(1)
	}

	// Flush changes and restart Docker
	cmd := exec.Command("systemctl", "daemon-reload")
	err := cmd.Run()
	if err != nil {
		log.Println(err)
		exit(1)
	}
	cmd = exec.Command("systemctl", "restart", "docker")
	err = cmd.Run()
	if err != nil {
		log.Println(err)
		exit(1)
	}
}

func save_settings(command WebSocketCommand) WebSocketCommand {
	log.Println("Started Applying Docker Proxy Settings")

	// TODO: Generalize this for all settings
	var proxy_settings ProxySettings
	err := json.Unmarshal([]byte(command.Arguments[0]), &proxy_settings)
	if err != nil {
		log.Println("Failed to de-serialize the JSON String!")
		exit(1)
	}

	apply_proxy_settings(proxy_settings)
	
	// TODO: Error handling if settings fail to apply
	command.Response = append(command.Response, "0")
	return command
}

func apply_proxy_settings(proxy_settings ProxySettings) {
	// Checks if File /etc/sysconfig/docker exists
	if _, err := os.Stat("/etc/sysconfig/docker"); err == nil {
		update_sysconfig(proxy_settings)
	} else {
		update_systemd(proxy_settings)
	}

	// Setting up Environment Variables
	os.Setenv("https_proxy", proxy_settings.HTTPS)
	os.Setenv("http_proxy", proxy_settings.HTTP)
	os.Setenv("ftp_proxy", proxy_settings.FTP)
	os.Setenv("no_proxy", proxy_settings.NO)

	log.Println("Applied Docker Proxy Settings")
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
