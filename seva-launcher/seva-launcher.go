package main

import (
	"embed"
	"encoding/json"
	"errors"
	"flag"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"strings"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/melbahja/got"
	"github.com/skratchdot/open-golang/open"
)

var store_url = "https://raw.githubusercontent.com/StaticRocket/seva-apps/main"
var addr = flag.String("addr", "0.0.0.0:8000", "http service address")
var no_browser = flag.Bool("no-browser", false, "do not launch browser")
var upgrader = websocket.Upgrader{}
var container_id_list [2]string


//go:embed web/*
var content embed.FS

type Containers []struct {
	ID       string `json:"ID"`
	Name     string `json:"Name"`
	Command  string `json:"Command"`
	Project  string `json:"Project"`
	Service  string `json:"Service"`
	State    string `json:"State"`
	Health   string `json:"Health"`
	ExitCode int    `json:"ExitCode"`
	Publishers []struct {
		URL           string `json:"URL"`
		TargetPort    int    `json:"TargetPort"`
		PublishedPort int    `json:"PublishedPort"`
		Protocol      string `json:"Protocol"`
	} `json:"Publishers"`
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
		switch(string(message)) {
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
		log.Fatal(err)
	}
	err = os.Chdir("/tmp/seva-launcher")
	if err != nil {
		log.Fatal(err)
	}
}

func launch_browser() {
	err := open.Start("http://localhost:8000/")
	if err != nil {
		log.Println("Host browser not detected. Fetching one through docker")
		// TODO
		//os.exec("docker run -it firefox")
	}
}

func start_design_gallery() {
	log.Println("Starting local design gallery service")
	cmd := exec.Command("docker", "run", "--rm", "-d", "-p", "8001:80", "ghcr.io/staticrocket/seva-design-gallery:latest")
	output, err := cmd.CombinedOutput()
	log.Printf("|\n%s\n", output)
	if err != nil {
		log.Fatal("Failed to start local design gallery container!")
	}
	container_id_list[0] = strings.TrimSpace(string(output))
}

func start_app() string {
	log.Println("Starting selected app")
	cmd := exec.Command("docker-compose", "-p", "seva-launcher", "up", "-d")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal("Failed to start selected app!")
	}
	output_s := strings.TrimSpace(string(output))
	log.Printf("|\n%s\n", output_s)
	return output_s
}

func stop_app() string {
	log.Println("Stopping selected app")
	cmd := exec.Command("docker-compose", "-p", "seva-launcher", "down", "--remove-orphans")
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
		log.Fatal(err)
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
			log.Fatal("Failed to remove old files")
		}
	}
	g := got.New()
	for _, element := range files {
		url := store_url + "/" + name + "/" + element
		log.Println("Fetching " + element + " from: " + url)
		err := g.Download(url, element)
		if err != nil {
			log.Fatal(err)
		}
	}
	return string("0")
}

func is_running(name string) string {
	log.Println("Checking if " + name + " is running")
	cmd := exec.Command("docker-compose", "-p", "seva-launcher", "ps", "--format", "json")
	output, err := cmd.Output()
	if err != nil {
		log.Fatal("Failed to check if app is running!")
	}
	var containers Containers
	err = json.Unmarshal([]byte(output), &containers)
	if err != nil {
		log.Fatal("Failed to parse JSON from docker-compose!")
	}
	for _, element := range containers {
		if element.Name == name {
			return string("1")
		}
	}
	return string("0")
}

func exit() {
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
	// TODO
}

func setup_exit_handler() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		exit()
		os.Exit(0)
	}()
}

func handle_requests() {
	router := mux.NewRouter()
	router.HandleFunc("/ws", echo)
	log.Println("Listening for websocket messages at " + *addr + "/ws")
	root_content, err := fs.Sub(content, "web")
	if err != nil {
		log.Fatal("No files to server for web interface!")
	}
	router.PathPrefix("/").Handler(http.FileServer(http.FS(root_content)))
	log.Fatal(http.ListenAndServe(*addr, router))
}


func main() {
	setup_exit_handler()

	flag.Parse()
	log.Println("Setting up working directory")
	setup_working_directory()
	start_design_gallery()

	if !*no_browser {
		log.Println("Launching browser")
		launch_browser()
	}

	handle_requests()
}
