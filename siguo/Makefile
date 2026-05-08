GO ?= /opt/homebrew/bin/go
BIN ?= bin/siguo

.PHONY: test build run docker web codegen clean

test:
	$(GO) test ./...

build:
	mkdir -p bin
	$(GO) build -o $(BIN) ./cmd/siguo

run:
	$(GO) run ./cmd/siguo

web:
	mkdir -p web/dist internal/static/dist
	cp assets/map3.jpg assets/map8.jpg web/dist/
	cp web/dist/* internal/static/dist/
	@echo "web/dist is dependency-free and mirrored into internal/static/dist"

codegen:
	@echo "protocol types are hand-authored for the dependency-free v1"

docker:
	docker build -t siguo:latest .

clean:
	rm -rf bin
