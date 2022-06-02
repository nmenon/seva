# Server Configuration

These are services and configs to be run on the embedded platform.

## Services
### server.py

Main websocket interface. Required for functionality. Systemd service coming soon.

## Configuration files
### lighttpd.conf

An example lighttpd configuration file. This assumes the following project structure:
```
/opt/seva/
|-- lighttpd.conf
|-- web
|   |-- assets
|   |   `-- ...
|   `-- ...
`-- websocket
    |-- server.py
    `-- seva-docker
        |-- docker-compose.yml
        `-- metadata.json
```
