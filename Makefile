.PHONY: help image testing
.SILENT: 
.DEFAULT_GOAL = help

PROJECT_NAME = $(shell basename $(CURDIR))

TARGET_OS = "windows" "linux" "darwin" "openbsd"
TARGET_ARCH = "amd64" "arm64" 

ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
$(eval $(ARGS):;@:)

clean:
	rm -rf $(CURDIR)/build/*

cross-build: clean vendor  ## Build project for all supported platform
	for os in $(TARGET_OS);do                                                                                                  \
		for arch in $(TARGET_ARCH);do                                                                                          \
			echo $$os/$$arch;                                                                                                  \
			env GOOS=$$os GOARCH=$$arch go build -o $(CURDIR)/build/docker-healthcheck-$$os-$$arch ./cmd/healthcheck/main.go ; \
		done;                                                                                                                  \
	done;

build: clean vendor ## Build project for local
	go build -o $(CURDIR)/build/docker-healthcheck ./cmd/healthcheck/main.go

install: build ## Install docker plugin
	mkdir -p /usr/lib/docker/cli-plugins
	cp $(CURDIR)/build/docker-healthcheck /usr/lib/docker/cli-plugins/docker-healthcheck
	chmod -R 700 /usr/lib/docker/cli-plugins/*

run: ## Run without build project
	go run ./cmd/healthcheck/main.go

docker-build: ## Build project in container
	docker build --force-rm -t healthcheck/local -f Dockerfile .

docker: docker-build ## Run project in container
	docker run -ti --name healthcheck --rm -v /var/run/docker.sock:/var/run/docker.sock --privileged healthcheck/local

docker-build-debug: ## Build project in debug container
	docker build --force-rm -t healthcheck/debug -f Dockerfile.debug .

docker-debug: ## Run project in debug container
	docker run -ti --name healthcheck --rm \
		-v $$(pwd):/tmp/healthcheck \
		-v build:/tmp/healthcheck/build \
		-v /var/run/docker.sock:/var/run/docker.sock --privileged healthcheck/debug

vendor:
	go mod tidy

help: ## Display all commands available
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'