### Change this variables ###

NAME=market-manager

VERSION ?= "dev"

REPO=github.com/dohernandez/market-manager

### Change this variables


# Packages
PACKAGES =$(shell go list ./pkg/...)

# Build
BUILD_DIR ?= build
GO_LINKER_FLAGS=-ldflags="-s -w -X main.version=$(VERSION)"
BINARY=${NAME}
BINARY_SRC=$(REPO)

# Colorz
NO_COLOR=\033[0m
OK_COLOR=\033[32;01m
ERROR_COLOR=\033[31;01m
WARN_COLOR=\033[33;01m

# Filters variables
CFLAGS=-g
export CFLAGS

all: usage

usage:
	@echo "---------------------------------------------"
	@echo "Usage:"
	@echo "  usage 						 - Shows this dialog."
	@echo " "
	@echo " Working with native GO"
	@echo "  run 							 - Runs targets to simply make it work"
	@echo "  deps 							 - Ensures dependencies using dep and installs several required tools"
	@echo "  deps-dev 						 - Install dependencies required only for development"
	@echo "  build							 - Buld the application"
	@echo "  build-docs						 - Build API documentation from RAML files"
	@echo "  install 						 - Install app using go install"
	@echo "  test 							 - Run all tests"
	@echo "  test-unit 						 - Run only unit tests"
	@echo "  test-integration 				 - Run integration unit tests"
	@echo "  dev-run-http 						 - Run REST API"
	@echo "  dev-migrate 						 - Run migrations"
	@echo " "
	@echo " Working with docker containers"
	@echo "  dev-docker-start					 - Start docker containers"
	@echo "  dev-docker-stop 					 - Stop docker containers"
	@echo "  dev-docker-deps 					 - Install dependencies using docker container"
	@echo "  dev-docker-migration 					 - Run migration using docker container"
	@echo "  dev-docker-test-unit 					 - Run all unit tests using docker container"
	@echo "  dev-docker-test-integration 			 - Run all integration tests using docker container"
	@echo "  dev-docker-logs [<CONTAINER>] 			 - Print container logs"
	@echo " "
	@exit 0

run: clean deps build install

clean:
	@printf "$(OK_COLOR)==> Cleaning build artifacts$(NO_COLOR)\n"
	@rm -rf $(BUILD_DIR)

# Deps
deps:
	@git config --global url."https://${GITHUB_TOKEN}@github.com/dohernandez/".insteadOf "https://github.com/dohernandez/"
	@git config --global http.https://gopkg.in.followRedirects true

	@printf "$(OK_COLOR)==> Installing dep$(NO_COLOR)\n"
	@go get -u github.com/golang/dep/cmd/dep

	@printf "$(OK_COLOR)==> Ensuring dependencies$(NO_COLOR)\n"
	@dep ensure

deps-dev: deps-dev-overalls
	@printf "$(OK_COLOR)==> Installing CompileDaemon$(NO_COLOR)\n"
	@go get github.com/githubnemo/CompileDaemon

deps-dev-overalls:
	@printf "$(OK_COLOR)==> Installing overalls(NO_COLOR)\n"
	@go get github.com/go-playground/overalls

# Build
build:
	@printf "$(OK_COLOR)==> Building Binary $(NO_COLOR)\n"
	@go build -o ${BUILD_DIR}/${BINARY} ${GO_LINKER_FLAGS} ${BINARY_SRC}

build-docs:
	@docker run --rm -w "/data/" -v `pwd`:/data mattjtodd/raml2html:7.0.0 raml2html  -i "docs/raml/api.raml" -o "docs/api.html"

# Install
install:
	@printf "$(OK_COLOR)==> Installing using go install$(NO_COLOR)\n"
	@go install ${REPO}

codecov: deps-dev-overalls
	@printf "$(OK_COLOR)==> Running code coverage $(NO_COLOR)\n"
	@overalls -project=github.com/hellofresh/market-manager -ignore="adapter,vendor,.glide,common" -covermode=count

# Test
test: test-unit test-integration

test-unit:
	@printf "$(OK_COLOR)==> Running unit tests$(NO_COLOR)\n"
	@go test -race $(PACKAGES)

test-integration:
	@printf "$(OK_COLOR)==> Running integration tests$(NO_COLOR)\n"
	@go test -godog -stop-on-failure

# Dev
dev-run-http:
	@CompileDaemon -build="make install" -graceful-kill -command="market-manager http"

dev-migrate:
	@docker-compose run --rm --name app-migrations base market-manager migrate up

# Dev with docker
dev-docker-start:
	@printf "$(OK_COLOR)==> Starting docker containers$(NO_COLOR)\n"
	@docker-compose up -d

dev-docker-stop:
	@printf "$(OK_COLOR)==> Stopping docker containers$(NO_COLOR)\n"
	@docker-compose down

dev-docker-deps:
	@printf "$(OK_COLOR)==> Installing dependencies using docker container$(NO_COLOR)\n"
	@docker-compose exec http make deps
	@docker-compose exec http make deps-dev

dev-docker-migration:
	@printf "$(OK_COLOR)==> Running migration using docker container$(NO_COLOR)\n"
	@docker-compose exec http market-manager migrate up

dev-docker-test-unit:
	@printf "$(OK_COLOR)==> Running unit test using docker container$(NO_COLOR)\n"
	@printf "Don't forget before run unit test to update deps\n"
	@docker-compose exec http make test-unit

dev-docker-test-integration:
	@printf "$(OK_COLOR)==> Running integration test using docker container$(NO_COLOR)\n"
	@printf "Don't forget before run integration test to update deps and run migration if need it\n"
	@docker-compose exec http make test-integration

dev-docker-logs:
	@docker-compose logs -f ${CONTAINER}

.PHONY: all usage run clean deps deps-dev build build-docs install test test-unit test-integration \
dev-run-consumer dev-run-http dev-migrate dev-docker-start dev-docker-stop dev-docker-build dev-docker-migration \
dev-docker-test-integration dev-docker-logs