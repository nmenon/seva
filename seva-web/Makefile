build/web: $(shell find . -name '*.dart')
	flutter build web --web-renderer html

.PHONY: all seva-web clean
all: seva-web
seva-web: build/web
clean:
	flutter clean
