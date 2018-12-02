### --------------------------------------------------------------------------------------------------------------------
### Variables
### (https://www.gnu.org/software/make/manual/html_node/Using-Variables.html#Using-Variables)
### --------------------------------------------------------------------------------------------------------------------
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
	@echo "  usage 					- Shows this dialog."
	@echo " "
	@echo " Working with native GO"
	@echo "  run 						- Runs targets to simply make it work"
	@echo "  deps 						- Ensures dependencies using dep and installs several required tools"
	@echo "  deps-dev 					- Install dependencies required only for development"
	@echo "  build						- Buld the application"
	@echo "  build-docs					- Build API documentation from RAML files"
	@echo "  install 					- Install app using go install"
	@echo "  test 						- Run all tests"
	@echo "  test-unit 					- Run only unit tests"
	@echo "  test-integration 				- Run integration unit tests"
	@echo "  dev-run-http 					- Run REST API"
	@echo "  dev-migrate 					- Run migrations"
	@echo " "
	@echo " Working with docker containers"
	@echo "  dev-docker-start				- Start docker containers"
	@echo "  dev-docker-stop 				- Stop docker containers"
	@echo "  dev-docker-deps 				- Install dependencies using docker container"
	@echo "  dev-docker-migration 			   	- Run migration using docker container"
	@echo "  dev-docker-test-unit 				- Run all unit tests using docker container"
	@echo "  dev-docker-test-integration 			- Run all integration tests using docker container"
	@echo "  dev-docker-logs [<CONTAINER>] 		- Print container logs"
	@echo " "
	@echo " Tools"
	@echo "  fix-style					- Removes blank lines (grouping) from imports in project files and applies goimports (with gofmt)  "
	@echo " "
	@echo " Arguments"
	@echo "  TAGS 						Optional tag(s) to run. Filter scenarios by tags. Default \"~@notImplemented\""
	@echo "  						  - \"@dev\": run all scenarios with wip tag"
	@echo "  						  - \"~@notImplemented\": exclude all scenarios with wip tag"
	@echo "  						  - \"@dev && ~@notImplemented\": run wip scenarios, but exclude new"
	@echo "  						  - \"@dev,@undone\": run wip or undone scenarios"
	@echo "  CONTAINER 					Container to look at the logs"
	@echo " "
	@echo " Examples"
	@echo " - To run only scenarios:"
	@echo "    make dev-integration-run TAGS=@dev"
	@exit 0

run: clean deps build install

#-----------------------------------------------------------------------------------------------------------------------
# House keeping - Cleans our project: deletes binaries
#-----------------------------------------------------------------------------------------------------------------------
clean:
	@printf "$(OK_COLOR)==> Cleaning build artifacts$(NO_COLOR)\n"
	@rm -rf $(BUILD_DIR)

### --------------------------------------------------------------------------------------------------------------------
# Dependencies
### --------------------------------------------------------------------------------------------------------------------
deps:
	@git config --global url."https://${GITHUB_TOKEN}@github.com/dohernandez/".insteadOf "https://github.com/dohernandez/"
	@git config --global http.https://gopkg.in.followRedirects true

	@printf "$(OK_COLOR)==> Installing dep$(NO_COLOR)\n"
	@go get -u github.com/golang/dep/cmd/dep

	@printf "$(OK_COLOR)==> Ensuring dependencies$(NO_COLOR)\n"
	@dep ensure -v

deps-dev: deps-dev-overalls
	@printf "$(OK_COLOR)==> Installing CompileDaemon$(NO_COLOR)\n"
	@go get github.com/githubnemo/CompileDaemon

deps-dev-overalls:
	@printf "$(OK_COLOR)==> Installing overalls$(NO_COLOR)\n"
	@go get github.com/go-playground/overalls

### --------------------------------------------------------------------------------------------------------------------
# Building
### --------------------------------------------------------------------------------------------------------------------
build:
	@printf "$(OK_COLOR)==> Building Binary $(NO_COLOR)\n"
	go build -o ${BUILD_DIR}/${BINARY} ${GO_LINKER_FLAGS} ${BINARY_SRC}/cmd/${BINARY}

build-docs:
	@docker run --rm -w "/data/" -v `pwd`:/data mattjtodd/raml2html:7.0.0 raml2html  -i "docs/raml/api.raml" -o "docs/api.html"

### --------------------------------------------------------------------------------------------------------------------
# Installing
### --------------------------------------------------------------------------------------------------------------------
install:
	@printf "$(OK_COLOR)==> Installing using go install$(NO_COLOR)\n"
	@go install ${REPO}/cmd/${BINARY}

codecov: deps-dev-overalls
	@printf "$(OK_COLOR)==> Running code coverage$(NO_COLOR)\n"
	@overalls -project=github.com/dohernandez/market-manager -ignore="adapter,vendor,.glide,common" -covermode=count

### --------------------------------------------------------------------------------------------------------------------
# Testing
### --------------------------------------------------------------------------------------------------------------------
test: test-unit test-integration

test-unit:
	@printf "$(OK_COLOR)==> Running unit tests$(NO_COLOR)\n"
	@go test -race $(PACKAGES)

test-integration:
	@printf "$(OK_COLOR)==> Running integration tests$(NO_COLOR)\n"
	@go test -godog -stop-on-failure -tag="${TAGS}" -feature="${FEATURE}"

### --------------------------------------------------------------------------------------------------------------------
# Developing
### --------------------------------------------------------------------------------------------------------------------
dev-run-http:
	@CompileDaemon -build="make install" -graceful-kill -command="market-manager http"

dev-run-scheduler:
	@CompileDaemon -build="make install" -graceful-kill -command="market-manager scheduler"

dev-migrate:
	@docker-compose run --rm --name app-migrations base market-manager migrate up

dev-fix-style:
	@./resources/dev/fix-style.sh

### --------------------------------------------------------------------------------------------------------------------
# Dev with docker
### --------------------------------------------------------------------------------------------------------------------
dev-docker-start:
	@printf "$(OK_COLOR)==> Starting docker containers$(NO_COLOR)\n"
	@docker-compose -f docker-compose.yml -f ./resources/dev/docker-compose.dev.yml -p market-manager-test up -d

dev-docker-stop:
	@printf "$(OK_COLOR)==> Stopping docker containers$(NO_COLOR)\n"
	@docker-compose -f docker-compose.yml -f ./resources/dev/docker-compose.dev.yml -p market-manager-test down

dev-docker-deps:
	@printf "$(OK_COLOR)==> Installing dependencies using docker container$(NO_COLOR)\n"
	@docker-compose -f docker-compose.yml -f ./resources/dev/docker-compose.dev.yml -p market-manager-test exec http make deps
	@docker-compose -f docker-compose.yml -f ./resources/dev/docker-compose.dev.yml -p market-manager-test exec http make deps-dev

dev-docker-migration:
	@printf "$(OK_COLOR)==> Running migration using docker container$(NO_COLOR)\n"
	@docker-compose -f docker-compose.yml -f ./resources/dev/docker-compose.dev.yml -p market-manager-test exec http market-manager migrate up

dev-docker-test-unit:
	@printf "$(OK_COLOR)==> Running unit test using docker container$(NO_COLOR)\n"
	@printf "Don't forget before run unit test to update deps\n"
	@docker-compose -f docker-compose.yml -f ./resources/dev/docker-compose.dev.yml -p market-manager-test exec http make test-unit

dev-docker-test-integration:
	@printf "$(OK_COLOR)==> Running integration test using docker container$(NO_COLOR)\n"
	@printf "Don't forget before run integration test to update deps and run migration if need it\n"
	@docker-compose -f docker-compose.yml -f ./resources/dev/docker-compose.dev.yml -p market-manager-test exec http make test-integration TAGS=${TAGS} FEATURE=${FEATURE}

dev-docker-logs:
	@docker-compose -f docker-compose.yml -f ./resources/dev/docker-compose.dev.yml -p market-manager-test logs -f ${CONTAINER}

dev-docker-bash:
	@docker-compose -f docker-compose.yml -f ./resources/dev/docker-compose.dev.yml -p market-manager-test exec ${CONTAINER} bash

### --------------------------------------------------------------------------------------------------------------------
# Prod with docker
### --------------------------------------------------------------------------------------------------------------------
prod-docker-start:
	@printf "$(OK_COLOR)==> Starting docker containers$(NO_COLOR)\n"
	@docker-compose -p market-manager up -d

prod-docker-stop:
	@printf "$(OK_COLOR)==> Stopping docker containers$(NO_COLOR)\n"
	@docker-compose -p market-manager down

prod-docker-migration:
	@printf "$(OK_COLOR)==> Running migration using docker container$(NO_COLOR)\n"
	@docker-compose -p market-manager exec http market-manager migrate up

prod-docker-logs:
	@docker-compose -p market-manager logs -f ${CONTAINER}

prod-docker-bash:
	@docker-compose -p market-manager exec ${CONTAINER} bash

### --------------------------------------------------------------------------------------------------------------------
### RULES
### (https://www.gnu.org/software/make/manual/html_node/Rule-Introduction.html#Rule-Introduction)
### --------------------------------------------------------------------------------------------------------------------

.PHONY: all usage run clean deps deps-dev build build-docs install test test-unit test-integration \
dev-run-consumer dev-run-http dev-migrate dev-docker-start dev-docker-stop dev-docker-build dev-docker-migration \
dev-docker-test-integration dev-docker-logs dev-fix-style
