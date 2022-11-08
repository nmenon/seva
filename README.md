# Seva

Embedded platform demo tool using Flutter and Go. This service attempts to act
as a remote head for embedded devices, allowing users to easily launch approved
demos from a predetermined source using docker-compose.

## Installation

Download the static seva-launcher binary from the release page or see the below
build instructions:

### Seva-Web

A Makefile as been included for ease of use:
```
cd seva-web
make
```

### Seva-Launcher

A Makefile as been included for ease of use:
```
cd seva-launcher
make
```

## Usage

The tool tries to take care of most of the heavy lifting but there are some
assumptions we have to make about the environment the tool is running in. Right
now the current assumptions are:

 1. Docker is already installed
 2. The user running seva-launcher is in the docker group or has been granted
    permission to interact with the docker socket
 3. The target platform is Linux based

Demos may make further assumptions and these will need to be advertised in the
associated store. For instance, the app_thermostat_demo right now is higly
specialized around arm64/aarch64 and, as such, can not be executed on
amd64/x86_64.

Access the web interface using either:

- The on device interface with `http://localhost/`
- The remote interface with `http://<device-ip-or-hostname>/`

The store page is currently set to use the locally hosted design gallery page
provided by the [seva-design-gallery
container](https://github.com/StaticRocket/seva-design-gallery) on port 8001.
While this can be accessed without being launched from the Seva Control Center,
it will not behave correctly for the following reason.

### Details

Seva Control Center is not only a web interface for seva-launcher but it also
acts as a sort of reverse proxy, allowing the store page to pass a single
filtered message through it, back to seva-launcher to load the specified apps.
That being said, loading apps is the only functionality it has access to and
even that relies on:

 1. The store page has to be launched through the Seva Control Pannel
 2. The store responding to the seva-init handshake routine

These two conditions resort in a communication band good enough to get around
most brower's same-origin policy regarding postMessages. As an added measure
though, Seva Control Center doesn't blindly pass along messages to
seva-launcher. The only postMessages it's allowed to receive are application
names, which it checks against a known store for. If the application isn't
found, the command is ignored. If it is found then the command is send to
seva-launcher instructing it to fetch the associated files.

The result of this fetch is then relayed back to the web interface through a
websocket at `http://localhost:8000/ws`. Seva Control Center can only interact
with seva-launcher in this manner.

### Why is this so complicated?

Using it really isn't, but hacking it can be. Completely decoupling the store
from the web interface and the controller allows all components to be modified
freely and distributed in a verity of ways.

 - Need a fully self-contained touch screen compatible interface for demos? Way
   ahead of you. We even have a preconfigured Firefox container just for it.
 - Need a demo running on a device without the overhead of a GUI? Cool, we
   already support that, the GUI runs entirely in the accessors web browser if
   you want it, but you could always just hook into the websocket API directly,
   too.
 - Want to host your own storefront? Go right ahead. See the old
   [seva-store](https://github.com/StaticRocket/seva-store) for an example of
   the minimum requirements for interacting with the Seva Control Center and
   [seva-apps](https://github.com/StaticRocket/seva-apps) for an example of the
   app repo structure. (Though right now you'll have to change the store
   address that's compiled into the web interface and launcher. We're looking
   into making that more friendly.)

## Future plans and features

I may not be the one to implement these, but I'd love to see them in place.

 1. Proxy settings. We've already got some wonderful work for this, but there's
    a little more left before it's ready.
 2. Multi-store support. Allow multiple stores to be configured through the web
    interface.
 3. Local app cacheing. Currently the service clobbers the last running app and
    it's files. We should be able to keep them on disk and switch apps easily
    through keeping a full directory tree.
 4. Integration with a web VNC client like [Apache
    Guacamole](https://guacamole.apache.org/) for headless/remote graphical
    demos.
 5. Integration with web based terminal emulators like
    [ttyd](https://tsl0922.github.io/ttyd/)
 6. Integration with web based text editors / IDEs. Like Atom / VS Code /
    Jupyter Notebook and the likes.
 7. Integration with external learning platforms.
