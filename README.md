# Seva

Embedded platform demo tool using flutter. This service attempts to act as a remote head for embedded devices, allowing users to easily launch approved demos from a predetermined source using docker-compose.

## Installation

Build the project with flutter:
```
flutter build web
```

Setup the following directory structure on the embedded platform:
```
/opt/seva/
|-- web
`-- websocket
```

Copy `build/web/*` to `/opt/seva/web/` on the target device.

Configure your web server of chice and make sure the files can be read by that process. An example configuration for lighttpd is provided [here](server/lighttpd.conf).

Copy the python server script `server/server.py` to `/opt/seva/websocket/`. Configure that process to start at boot with your preferred init system.

Access the web interface either:
- On device with http://localhost/
- Remotely with http://<device-ip-or-hostname>/
