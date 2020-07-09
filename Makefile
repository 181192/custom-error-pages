PROJECTNAME=$(shell basename "$(PWD)")
CURRENT=$(shell echo $(PWD))

.PHONY: help

help: Makefile
	@echo
	@echo " Choose a command run in "$(PROJECTNAME)":"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo

## install: Install missing dependencies.
install:
	go mod download

## run: Runs the application.
run:
	ERROR_FILES_PATH=$(CURRENT)/www go run ./

## build: Builds the project.
build: main.go
	go build -o build/custom-error-pages

## build-all: Build all linux plattforms
build-all: main.go
	for arch in amd64; do \
			for os in linux; do \
				CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch go build -o "build/custom_error_pages"$$os"_$$arch" $(LDFLAGS) -ldflags "-X 'github.com/181192/ops-cli/pkg/util/version.Version=$$(git describe --tags --abbrev=0)' -X 'github.com/181192/ops-cli/pkg/util/version.GitCommit=$$(git rev-parse --short HEAD)'"; \
			done; \
		done;
		for arch in arm arm64; do \
			for os in linux; do \
				CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch go build -o "build/custom_error_pages"$$os"_$$arch" $(LDFLAGS) -ldflags "-X 'github.com/181192/ops-cli/pkg/util/version.Version=$$(git describe --tags --abbrev=0)' -X 'github.com/181192/ops-cli/pkg/util/version.GitCommit=$$(git rev-parse --short HEAD)'"; \
			done; \
	done;

## docker-build: Build docker image
docker-build:
	docker build -t custom-error-pages .

## docker-run: Run docker image
docker-run: docker-build
	docker run --rm -p 8080:8080 custom-error-pages:latest

k3d-import: docker-build
	k3d i custom-error-pages

k8s-deploy: k3d-import
	kubectl apply -k k8s/nginx-ingress
	kubectl rollout restart deployment nginx-ingress-default-backend -n kube-system
