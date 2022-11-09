package main

import (
	"embed"
	"flag"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"strings"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/skratchdot/open-golang/open"
)

var store_url = "https://raw.githubusercontent.com/StaticRocket/seva-apps/main"
var addr = flag.String("addr", "0.0.0.0:8000", "http service address")
var no_browser = flag.Bool("no-browser", false, "do not launch browser")
var docker_browser = flag.Bool("docker-browser", false, "force use of docker browser")
var container_id_list [2]string
var docker_compose string

//go:embed web/*
var content embed.FS

//go:embed docker-compose
var docker_compose_bin []byte

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
		err := open.Start("http://localhost:8000/#/")
		if err != nil {
			log.Println("Host browser not detected, fetching one through docker")
			go launch_docker_browser()
		}
	}
}

func launch_docker_browser() {
	xdg_runtime_dir := os.Getenv("XDG_RUNTIME_DIR")
	user, _ := user.Current()
	output := docker_run("--rm", "--privileged", "--network", "host",
		"-v", "/tmp/.X11-unix",
		"-e", "XAUTHORITY",
		"-e", "XDG_RUNTIME_DIR=/tmp",
		"-e", "DISPLAY",
		"-e", "WAYLAND_DISPLAY",
		"-e", "https_proxy",
		"-e", "http_proxy",
		"-e", "no_proxy",
		"-v", xdg_runtime_dir+":/tmp",
		"--user="+user.Uid+":"+user.Gid,
		"ghcr.io/staticrocket/seva-browser:latest",
		"http://localhost:8000/#/",
	)
	output_strings := strings.Split(strings.TrimSpace(string(output)), "\n")
	container_id_list[1] = output_strings[len(output_strings)-1]
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
	output := docker_run("--rm", "-p", "8001:80",
		"ghcr.io/staticrocket/seva-design-gallery:latest",
	)
	output_strings := strings.Split(strings.TrimSpace(string(output)), "\n")
	container_id_list[0] = output_strings[len(output_strings)-1]
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
	router.HandleFunc("/ws", websocket_controller)
	log.Println("Listening for websocket messages at " + *addr + "/ws")
	root_content, err := fs.Sub(content, "web")
	if err != nil {
		log.Println("No files to server for web interface!")
		exit(1)
	}
	router.PathPrefix("/").Handler(http.FileServer(http.FS(root_content)))
	log.Println(http.ListenAndServe(*addr, router))
}

func check_env_vars() {
	for _, element := range []string{"DISPLAY", "WAYLAND_DISPLAY"} {
		env_var := os.Getenv(element)
		if len(env_var) > 0 {
			return
		}
	}
	log.Println("Environment variable DISPLAY or WAYLAND_DISPLAY must be set!")
	exit(1)
}

func main() {
	setup_exit_handler()
	check_env_vars()
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
