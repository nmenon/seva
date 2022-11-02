package main

import (
	"embed"
	"encoding/json"
	"errors"
	"flag"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/melbahja/got"
	"github.com/skratchdot/open-golang/open"
)

var store_url = "https://raw.githubusercontent.com/StaticRocket/seva-apps/main"
var addr = flag.String("addr", "0.0.0.0:8000", "http service address")
var no_browser = flag.Bool("no-browser", false, "do not launch browser")
var docker_browser = flag.Bool("docker-browser", false, "force use of docker browser")
var upgrader = websocket.Upgrader{}
var container_id_list [2]string
var docker_compose string

//go:embed web/*
var content embed.FS

//go:embed docker-compose
var docker_compose_bin []byte

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

func is_docker_compose_installed() bool {
	cmd := exec.Command("docker-compose", "-v")
	_, err := cmd.CombinedOutput()
	if err != nil {
		log.Println("Docker-compose is either not installed or cannot be executed")
		log.Println(err)
		log.Println("Using local install for now")
		return false
	}
	return true
}

func prepare_compose() string {
	if !is_docker_compose_installed() {
		ioutil.WriteFile("docker-compose", docker_compose_bin, 0755)
		return "./docker-compose"
	}
	return "docker-compose"
}

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		var resp = string("")
		switch string(message) {
		case "start_app":
			resp = start_app()
		case "load_app":
			var name []byte
			_, name, err = c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				break
			}
			resp = load_app(string(name))
		case "stop_app":
			resp = stop_app()
		case "get_app":
			resp = get_app()
		case "is_running":
			var name []byte
			_, name, err = c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				break
			}
			resp = is_running(string(name))
		default:
			resp = "Ignoring invalid command"
			log.Println(resp)
		}
		if resp != "" {
			err = c.WriteMessage(websocket.TextMessage, []byte(resp))
			if err != nil {
				log.Println("write:", err)
				break
			}
		}
	}
}

func setup_working_directory() {
	err := os.MkdirAll("/tmp/seva-launcher", os.ModePerm)
	if err != nil {
		log.Println(err)
		exit(1)
	}
	err = os.Chdir("/tmp/seva-launcher")
	if err != nil {
		log.Println(err)
		exit(1)
	}
}

func launch_browser() {
	if *docker_browser {
		go launch_docker_browser()
	} else {
		err := open.Start("http://localhost:8000/")
		if err != nil {
			log.Println("Host browser not detected, fetching one through docker")
			go launch_docker_browser()
		}
	}
}

func launch_docker_browser() {
	xdg_runtime_dir := os.Getenv("XDG_RUNTIME_DIR")
	output := docker_run("--rm", "--privileged", "--network", "host",
		"-v", "/tmp/.X11-unix",
		"-e", "XAUTHORITY",
		"-e", "XDG_RUNTIME_DIR",
		"-e", "WAYLAND_DISPLAY",
		"-v", xdg_runtime_dir+":"+xdg_runtime_dir,
		"ghcr.io/nmenon/demo_baseline_browser:latest",
		"http://localhost:8000/",
	)
	container_id_list[1] = strings.TrimSpace(string(output))
}

func docker_run(args ...string) []byte {
	args = append([]string{"run", "-d"}, args...)
	cmd := exec.Command("docker", args...)
	output, err := cmd.CombinedOutput()
	log.Printf("|\n%s\n", output)
	if err != nil {
		log.Println("Failed to start container!")
		log.Println(err)
		exit(1)
	}
	return output
}

func start_design_gallery() {
	log.Println("Starting local design gallery service")
	output := docker_run("--rm", "-p", "8001:80", "ghcr.io/staticrocket/seva-design-gallery:latest")
	container_id_list[0] = strings.TrimSpace(string(output))
}

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
		return ""
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

func exit(num int) {
	log.Println("Stopping non-app containers")
	for _, container_id := range container_id_list {
		if len(container_id) > 0 {
			cmd := exec.Command("docker", "stop", container_id)
			output, err := cmd.CombinedOutput()
			if err != nil {
				log.Printf("Failed to stop container %s : \n%s", container_id, output)
			}
		}
	}
	os.Exit(num)
}

func setup_exit_handler() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		exit(0)
	}()
}

func handle_requests() {
	router := mux.NewRouter()
	router.HandleFunc("/ws", echo)
	log.Println("Listening for websocket messages at " + *addr + "/ws")
	root_content, err := fs.Sub(content, "web")
	if err != nil {
		log.Println("No files to server for web interface!")
		exit(1)
	}
	router.PathPrefix("/").Handler(http.FileServer(http.FS(root_content)))
	log.Println(http.ListenAndServe(*addr, router))
}

func main() {
	setup_exit_handler()
	flag.Parse()

	log.Println("Setting up working directory")
	setup_working_directory()
	docker_compose = prepare_compose()

	go start_design_gallery()

	if !*no_browser {
		log.Println("Launching browser")
		launch_browser()
	}

	handle_requests()
}
