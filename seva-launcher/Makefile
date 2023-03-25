ARCH ?= $(shell uname -m)

ifeq ($(ARCH),x86_64)
	GOARCH=amd64
else ifeq ($(ARCH),aarch64)
	GOARCH=arm64
else
	@echo Unsupported architecture
endif

seva-launcher: docker-compose web $(shell find . -name '*.go')
	GOARCH=$(GOARCH) go build -tags netgo .

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

web: $(shell find ../seva-web/build/web)
	rm -rf web
	cp -r ../seva-web/build/web/ .

.PHONY: all clean run
all: seva-launcher
clean:
	rm -f seva-launcher
	rm -f docker-compose
	rm -rf web
run: seva-launcher
	./seva-launcher
