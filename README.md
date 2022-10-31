# Seva

Embedded platform demo tool using flutter. This service attempts to act as a remote head for embedded devices, allowing users to easily launch approved demos from a predetermined source using docker-compose.

## Installation

### Seva-Web

Build the project with flutter:
```
flutter build web
```

### Seva-Launcher

Build the project with go:
```
go build .
```

## Usage

Access the web interface either:
- On device with `http://localhost/`
- Remotely with `http://<device-ip-or-hostname>/`
