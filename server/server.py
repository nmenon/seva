'''
Author: Randolph Sapp
Info: Websocket server to wrap docker-compose
'''

import json
import asyncio
import subprocess
import urllib.request
from websockets import serve
from pathlib import Path

WORKING_DIR = Path("seva-docker")
STORE_URL = "https://raw.githubusercontent.com/StaticRocket/seva-apps/main/{}/{}"

METADATA_PATH = WORKING_DIR.joinpath("metadata.json")
COMPOSE_PATH = WORKING_DIR.joinpath("docker-compose.yml")


async def commander(websocket):
    """
    Write response back
    """
    async for command in websocket:
        print(f">{command}")
        resp = None
        if command == "start_app":
            resp = start_app()
        if command == "load_app":
            print("Waiting for command arguments")
            resp = load_app(await websocket.recv())
        if command == "stop_app":
            resp = stop_app()
        if command == "get_app":
            resp = get_app()
        if command == "is_running":
            print("Waiting for command arguments")
            resp = is_running(await websocket.recv())
        if resp is not None:
            await websocket.send(resp)


def prepare_dir():
    """
    Prepare the working directory
    """
    WORKING_DIR.mkdir(exist_ok=True)


def fetch_file(url, path):
    """
    Fetch the file from url and store it at path
    """
    urllib.request.urlretrieve(url, path)


def load_app(name):
    """
    Load app from store
    """
    print(f"Loading {name} from store")

    # kill any running apps using old config
    stop_app()

    # clean the working directory
    for file in WORKING_DIR.iterdir():
        if file.is_file():
            file.unlink()

    # fetch required files for new app
    fetch_file(STORE_URL.format(name, "metadata.json"), METADATA_PATH)
    fetch_file(STORE_URL.format(name, "docker-compose.yml"), COMPOSE_PATH)
    return str(0)


def start_app():
    """
    Start the currently loaded app
    """
    print("Starting app")
    status = subprocess.run(["docker-compose", "up", "-d"], cwd=WORKING_DIR)
    return str(status.returncode)


def stop_app():
    """
    Stop the currently loaded app
    """
    print("Stopping app")
    status = subprocess.run(["docker-compose", "down"], cwd=WORKING_DIR)
    return str(status.returncode)


def get_app():
    """
    Report information about the currently loaded app
    """
    print("Fetching app metadata from disk")
    metadata = ""
    if METADATA_PATH.exists():
        with open(METADATA_PATH, "r") as file:
            metadata = file.read()
    return metadata


def is_running(app_name):
    """
    Report whether the given app is running
    """
    print(f"Checking if {app_name} is running")
    status = subprocess.run(
        ["docker-compose", "ps", "--format", "json"],
        cwd=WORKING_DIR,
        capture_output=True,
    )
    running_containers = json.loads(status.stdout)
    for container in running_containers:
        print(container.get("Name", ""))
        print(app_name)
        if container.get("Name", "") == app_name:
            return "1"
    return "0"


async def main():
    """
    Initialize the server
    """
    async with serve(commander, "0.0.0.0", 8000):
        await asyncio.Future()


if __name__ == "__main__":
    prepare_dir()
    asyncio.run(main())
