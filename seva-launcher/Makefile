ARCH ?= $(shell uname -m)

seva-launcher: docker-compose web
	go build -tags netgo .

docker-compose:
	curl -s https://api.github.com/repos/docker/compose/releases/latest \
	| grep docker-compose-linux-$(ARCH) \
	| cut -d : -f 2,3 \
	| tr -d \" \
	| wget -qi - || true
	sha256sum -c docker-compose-linux-$(ARCH).sha256
	rm docker-compose-linux-$(ARCH).sha256
	mv docker-compose-linux-$(ARCH) docker-compose
	chmod +x docker-compose

web:
	cp -r ../seva-web/build/web/ .

.PHONY: all clean
all: seva-launcher
clean:
	rm seva-launcher
	rm docker-compose
	rm -r web
